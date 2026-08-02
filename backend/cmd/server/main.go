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
	"devopsconecta/backend/internal/config"
	"devopsconecta/backend/internal/ids"
	"devopsconecta/backend/internal/store"
)

func main() {
	cfg := config.Load()
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

	gen := ids.NewGenerator(int64(cfg.SnowflakeNode))
	app := api.New(cfg, store.New(db), gen)

	srv := &http.Server{
		Addr:         cfg.Host + ":" + cfg.Port,
		Handler:      app.Handler(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("DevOps Conecta rodando em http://%s:%s", cfg.Host, cfg.Port)
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
