package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"karat-server/pkg/handler"
	"karat-server/pkg/manager/middleware"
	user_manager "karat-server/pkg/manager/user"
	user_repo "karat-server/pkg/manager/user/repository"
	"karat-server/pkg/provider"
	"karat-server/pkg/provider/google"
	"karat-server/postgres"
	"karat-server/storage"
)

func main() {
	var cfg Config
	if err := cfg.Load(); err != nil {
		log.Fatal(err)
	}
	if err := run(cfg); err != nil {
		log.Fatal(err)
	}
}

func run(cfg Config) error {
	if cfg.Config.Google.DevLogin {
		log.Print("dev login is enabled; do not expose this server publicly")
	}
	db, err := postgres.InitDB(&cfg.Config.DB)
	if err != nil {
		return err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	googleClient := google.NewClient(cfg.Config.Google)
	tokens := provider.NewProviderClient(googleClient)
	users := user_manager.NewUserPort(user_repo.NewRepository(db))
	auth := middleware.NewMiddleware(tokens, users)

	router := gin.Default()
	router.GET("/healthz", health(db))
	router.Use(auth.AuthMiddleware())
	images := storage.New(cfg.Config.R2)
	if err := handler.InitHandlers(router, db, auth, cfg.Config.DB.GetSchema(), tokens, cfg.Config.Shop.Timezone, images); err != nil {
		return err
	}

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}
	errCh := make(chan error, 1)
	go func() {
		log.Printf("listening on %s", server.Addr)
		errCh <- server.ListenAndServe()
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	}
}

func health(db *gorm.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		sqlDB, err := db.DB()
		if err != nil || sqlDB.PingContext(ctx.Request.Context()) != nil {
			ctx.JSON(http.StatusServiceUnavailable, gin.H{"ok": false})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"ok": true})
	}
}
