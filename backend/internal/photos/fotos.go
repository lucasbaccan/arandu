package photos

// As fotos dos participantes ficam salvas no banco como data URLs base64
// (coluna participants.photo) — a fonte da verdade. Este pacote materializa
// essas fotos em arquivos no disco e as serve como arquivos estáticos
// (GET /api/fotos/{id}), para que nenhum payload JSON (lista de respostas,
// snapshot da apresentação) precise carregar o base64.
//
// A primeira requisição a uma foto decodifica o data URL do banco e grava o
// arquivo (mais lenta); as seguintes são servidas direto do disco. A limpeza
// diária (RunCleanupLoop) remove arquivos não acessados há mais de maxAge —
// por padrão 7 dias — e a foto é rematerializada do banco na próxima vez que
// for pedida.

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	// URLPrefix é a rota que serve as fotos como arquivo.
	URLPrefix = "/api/fotos/"

	// DefaultMaxAge é a idade máxima dos arquivos de foto em disco antes de
	// serem apagados pela limpeza diária.
	DefaultMaxAge = 7 * 24 * time.Hour

	// ImageMaxAge é quanto tempo o navegador pode servir a foto direto do
	// cache antes de revalidar com o servidor (If-Modified-Since → 304).
	// Pelo menos 1h para aliviar o tráfego na apresentação (muitas fotos
	// sendo re-solicitadas), sem travar a atualização de uma foto trocada
	// por mais de uma hora.
	ImageMaxAge = time.Hour

	// CleanupInterval roda a limpeza uma vez por dia.
	CleanupInterval = 24 * time.Hour
)

// ErrNotFound indica que o participante não tem foto para servir.
var ErrNotFound = errors.New("photos: foto não encontrada")

// Store materializa e serve as fotos dos participantes em disco.
type Store struct {
	dir string
}

// New cria o store de fotos, garantindo que o diretório exista.
func New(dir string) (*Store, error) {
	if dir == "" {
		dir = "./data/photos"
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("photos: criar diretório %s: %w", dir, err)
	}
	return &Store{dir: dir}, nil
}

// URL devolve a URL pública da foto de um participante.
func (s *Store) URL(id int64) string {
	return URLPrefix + strconv.FormatInt(id, 10)
}

func (s *Store) path(id int64) string {
	return filepath.Join(s.dir, strconv.FormatInt(id, 10)+".img")
}

// Remove apaga o arquivo em disco de um participante. Deve ser chamado sempre
// que a foto no banco muda ou é removida, para a próxima requisição
// rematerializar a partir do novo valor.
func (s *Store) Remove(id int64) {
	if err := os.Remove(s.path(id)); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Printf("photos: remover arquivo do participante %d: %v", id, err)
	}
}

// Serve atende a requisição de uma foto. Se o arquivo já existe em disco,
// serve direto dele (caminho rápido, sem tocar no banco); senão, busca o data
// URL via load(), grava o arquivo e serve (primeira vez mais lenta). load
// devolve a foto crua do banco ("") quando o participante não tem foto.
func (s *Store) Serve(w http.ResponseWriter, r *http.Request, id int64, load func() (string, error)) error {
	path := s.path(id)
	if _, err := os.Stat(path); err == nil {
		serveImage(w, r, path, "")
		return nil
	}

	dataURL, err := load()
	if err != nil {
		return err
	}
	if dataURL == "" {
		return ErrNotFound
	}

	mime, data, err := decodeDataURL(dataURL)
	if err != nil {
		return fmt.Errorf("photos: decodificar foto do participante %d: %w", id, err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("photos: gravar foto do participante %d: %w", id, err)
	}
	serveImage(w, r, path, mime)
	return nil
}

// serveImage entrega o arquivo com cache de ImageMaxAge (1h): nesse período
// o navegador reusa a imagem sem revalidar; depois disso revalida
// (If-Modified-Since → 304) antes de usar, então uma foto trocada pelo
// organizador aparece em no máximo uma hora — e o base64 nunca trafega no
// JSON.
func serveImage(w http.ResponseWriter, r *http.Request, path, mime string) {
	if mime == "" {
		if f, err := os.Open(path); err == nil {
			buf := make([]byte, 512)
			if n, err := f.Read(buf); err == nil && n > 0 {
				mime = http.DetectContentType(buf[:n])
			}
			f.Close()
		}
	}
	if mime != "" {
		w.Header().Set("Content-Type", mime)
	}
	w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d, must-revalidate", int(ImageMaxAge.Seconds())))
	http.ServeFile(w, r, path)
}

// decodeDataURL separa o mime do payload base64 de uma data URL de imagem
// (ex: "data:image/jpeg;base64,Zm9vCg==").
func decodeDataURL(dataURL string) (mime string, data []byte, err error) {
	const prefix = "data:"
	if !strings.HasPrefix(dataURL, prefix) {
		return "", nil, errors.New("foto não é uma data URL")
	}
	rest := dataURL[len(prefix):]
	semi := strings.IndexByte(rest, ';')
	comma := strings.IndexByte(rest, ',')
	if semi < 0 || comma < 0 || semi > comma {
		return "", nil, errors.New("data URL malformada")
	}
	mime = rest[:semi]
	if !strings.HasPrefix(mime, "image/") {
		return "", nil, errors.New("foto não é uma imagem")
	}
	data, err = base64.StdEncoding.DecodeString(rest[comma+1:])
	if err != nil {
		return "", nil, err
	}
	return mime, data, nil
}

// CleanupOlderThan apaga os arquivos de foto cujo último acesso (mtime) é
// anterior a maxAge. As fotos apagadas são rematerializadas do banco na
// próxima vez que forem solicitadas.
func (s *Store) CleanupOlderThan(maxAge time.Duration) error {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("photos: listar diretório: %w", err)
	}
	cutoff := time.Now().Add(-maxAge)
	removed := 0
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Before(cutoff) {
			if err := os.Remove(filepath.Join(s.dir, e.Name())); err == nil {
				removed++
			}
		}
	}
	if removed > 0 {
		log.Printf("photos: limpeza removeu %d arquivo(s) mais antigos que %s", removed, maxAge)
	}
	return nil
}

// RunCleanupLoop roda a limpeza diária (CleanupOlderThan) uma vez por dia, e
// mais uma vez logo no início, até o contexto ser cancelado.
func (s *Store) RunCleanupLoop(ctx context.Context, maxAge time.Duration) {
	cleanup := func() {
		if err := s.CleanupOlderThan(maxAge); err != nil {
			log.Printf("photos: limpeza diária: %v", err)
		}
	}
	cleanup()
	ticker := time.NewTicker(CleanupInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			cleanup()
		}
	}
}
