package photos

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestURL(t *testing.T) {
	s, err := New(t.TempDir())
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := s.URL(42); got != "/api/photos/42" {
		t.Errorf("URL esperada /api/photos/42, got %q", got)
	}
}

func TestDecodeDataURL(t *testing.T) {
	cases := []struct {
		name    string
		in      string
		wantMime string
		wantErr bool
	}{
		{"jpeg", "data:image/jpeg;base64,Zm9vCg==", "image/jpeg", false},
		{"png", "data:image/png;base64,YmFy", "image/png", false},
		{"sem prefixo", "image/jpeg;base64,Zm9v", "", true},
		{"não é imagem", "data:text/plain;base64,Zm9v", "", true},
		{"malformada", "data:image/jpeg,Zm9v", "", true},
		{"base64 inválido", "data:image/jpeg;base64,@@@", "", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mime, data, err := decodeDataURL(tc.in)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("erro esperado, got mime=%q data=%q", mime, data)
				}
				return
			}
			if err != nil {
				t.Fatalf("decode inesperado: %v", err)
			}
			if mime != tc.wantMime {
				t.Errorf("mime esperado %q, got %q", tc.wantMime, mime)
			}
			if len(data) == 0 {
				t.Error("dados decodificados não deveriam estar vazios")
			}
		})
	}
}

func TestServeMaterializesAndServes(t *testing.T) {
	s, err := New(t.TempDir())
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	// Primeira chamada: não existe arquivo, load() devolve o data URL e o
	// arquivo é materializado.
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, s.URL(7), nil)
	loads := 0
	err = s.Serve(rec, req, 7, func() (string, error) {
		loads++
		return "data:image/jpeg;base64,Zm9vCg==", nil
	})
	if err != nil {
		t.Fatalf("Serve (1ª): %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status esperado 200, got %d", rec.Code)
	}
	if rec.Body.String() != "foo\n" {
		t.Errorf("conteúdo esperado 'foo\\n', got %q", rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "image/jpeg") {
		t.Errorf("content-type esperado image/jpeg, got %q", ct)
	}
	if _, err := os.Stat(s.path(7)); err != nil {
		t.Errorf("arquivo deveria ter sido materializado: %v", err)
	}

	// Segunda chamada: arquivo já existe, serve direto do disco sem tocar no
	// load() (caminho rápido da apresentação).
	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, s.URL(7), nil)
	if err := s.Serve(rec2, req2, 7, func() (string, error) {
		loads++
		return "", nil
	}); err != nil {
		t.Fatalf("Serve (2ª): %v", err)
	}
	if rec2.Code != http.StatusOK || rec2.Body.String() != "foo\n" {
		t.Errorf("segunda chamada: status %d, body %q", rec2.Code, rec2.Body.String())
	}
	if loads != 1 {
		t.Errorf("load() deveria rodar só na 1ª chamada, rodou %d vezes", loads)
	}
}

func TestServeWithoutPhoto(t *testing.T) {
	s, err := New(t.TempDir())
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, s.URL(1), nil)
	err = s.Serve(rec, req, 1, func() (string, error) { return "", nil })
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("esperado ErrNotFound, got %v", err)
	}
}

func TestRemoveDeletesFile(t *testing.T) {
	s, err := New(t.TempDir())
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := os.WriteFile(s.path(3), []byte("x"), 0o644); err != nil {
		t.Fatalf("gravar arquivo: %v", err)
	}
	s.Remove(3)
	if _, err := os.Stat(s.path(3)); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("arquivo deveria ter sido removido, stat err = %v", err)
	}
}

func TestCleanupOlderThan(t *testing.T) {
	s, err := New(t.TempDir())
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	oldFile := filepath.Join(s.dir, "1.img")
	newFile := filepath.Join(s.dir, "2.img")
	for _, f := range []string{oldFile, newFile} {
		if err := os.WriteFile(f, []byte("x"), 0o644); err != nil {
			t.Fatalf("gravar %s: %v", f, err)
		}
	}
	old := time.Now().Add(-10 * 24 * time.Hour)
	if err := os.Chtimes(oldFile, old, old); err != nil {
		t.Fatalf("envelhecer arquivo: %v", err)
	}

	if err := s.CleanupOlderThan(DefaultMaxAge); err != nil {
		t.Fatalf("CleanupOlderThan: %v", err)
	}
	if _, err := os.Stat(oldFile); !errors.Is(err, os.ErrNotExist) {
		t.Error("arquivo antigo deveria ter sido apagado")
	}
	if _, err := os.Stat(newFile); err != nil {
		t.Errorf("arquivo recente não deveria ser apagado: %v", err)
	}
}
