package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"devopsconecta/backend/internal/api"
	"devopsconecta/backend/internal/ids"
	"devopsconecta/backend/internal/live"
	"devopsconecta/backend/internal/photos"
	"devopsconecta/backend/internal/store"

	// docs registra o spec gerado por `swag init` (ver Makefile, alvo
	// "swagger") no pacote swag — é o que GET /api/docs/doc.json lê.
	_ "devopsconecta/backend/docs"
)

// @title       Arandu API
// @version     1.0
// @description API REST do Arandu (coleta de respostas, apresentação ao vivo e administração de eventos) — as mesmas rotas usadas pelo frontend embutido, prontas para chamar diretamente e criar/gerenciar recursos por API.
// @description
// @description Autenticação de organizador (cookieAuth): faça login em outra aba pelo próprio app (POST /api/conta/entrar) — o cookie de sessão é httpOnly, então não dá para colá-lo aqui manualmente, mas o navegador o envia sozinho nas chamadas desta página por estarem na mesma origem.
// @description Visitante ao vivo (liveViewerToken): obtenha o token em POST /api/publico/eventos/{id}/ao-vivo/entrar e cole em "Authorize" — o Swagger UI passa a anexar "?token=" nas chamadas das rotas públicas de ao-vivo.
//
// @license.name Uso interno
//
// @basePath /
//
// @securityDefinitions.apikey cookieAuth
// @in header
// @name Cookie
//
// @securityDefinitions.apikey liveViewerToken
// @in query
// @name token
func main() {
	cfg := resolveConfig()
	if cfg.JWTSecret == "dev-secret-change-me" {
		log.Println("AVISO: JWT_SECRET não configurado, usando segredo de desenvolvimento.")
	}

	db, err := store.Open(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("main: %v", err)
	}
	defer db.Close()

	if err := store.Migrate(db); err != nil {
		log.Fatalf("main: %v", err)
	}

	photoStore, err := photos.New(cfg.PhotosDir)
	if err != nil {
		log.Fatalf("main: %v", err)
	}

	gen := ids.NewGenerator(int64(cfg.SnowflakeNode))
	app := api.New(cfg, store.New(db), gen, live.NewManager(), photoStore)

	srv := &http.Server{
		Addr:        cfg.Host + ":" + cfg.Port,
		Handler:     app.Handler(),
		ReadTimeout: 10 * time.Second,
		// Sem WriteTimeout: a apresentação ao vivo mantém conexões SSE
		// abertas por horas; um timeout global derrubaria esses streams.
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Limpeza diária das fotos materializadas: apaga arquivos mais antigos
	// que 7 dias (rematerializados do banco na próxima requisição).
	go photoStore.RunCleanupLoop(ctx, photos.DefaultMaxAge)

	go func() {
		log.Printf("Arandu rodando em http://%s:%s", cfg.Host, cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("main: servidor: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Encerrando servidor...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
}
