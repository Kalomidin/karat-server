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

	"github.com/jackc/pgx/v5/pgxpool"

	"karat-server/internal/auth"
	"karat-server/internal/config"
	"karat-server/internal/httpapi"
	"karat-server/internal/migrate"
	"karat-server/internal/store"
)

func main() {
	log.SetFlags(log.LstdFlags)
	cfg, err := config.FromEnv()
	if err != nil {
		log.Fatal(err)
	}
	if cfg.DevLogin {
		log.Print("dev login is enabled; do not expose this server publicly")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	err = pool.Ping(pingCtx)
	cancel()
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	if err := migrate.Up(ctx, pool); err != nil {
		log.Fatal(err)
	}

	server := &http.Server{
		Addr:              cfg.Addr,
		Handler:           httpapi.New(store.NewPostgres(pool), auth.NewGoogle(cfg.GoogleClientIDs, cfg.DevLogin), cfg.ShopLocation),
		ReadHeaderTimeout: 5 * time.Second,
	}
	errCh := make(chan error, 1)
	go func() {
		log.Printf("listening on %s", cfg.Addr)
		errCh <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Fatal(err)
		}
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}
}
