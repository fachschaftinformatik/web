package api

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fachschaftinformatik/web/internal/api/handler"
	"github.com/fachschaftinformatik/web/internal/avatars"
	"github.com/fachschaftinformatik/web/internal/buckets"
	"github.com/fachschaftinformatik/web/internal/config"
	"github.com/fachschaftinformatik/web/internal/database"
	"github.com/fachschaftinformatik/web/internal/email"
	"github.com/go-chi/httplog/v2"
)

func Run() error {
	cfg := config.New()

	// Initialize structured logger
	logger := httplog.NewLogger("fachschaft-api", httplog.Options{
		JSON:             true,
		LogLevel:         slog.LevelInfo,
		Concise:          false,
		RequestHeaders:   true,
		MessageFieldName: "message",
		TimeFieldFormat:  time.RFC3339,
	})

	if err := database.Migrate(cfg.DatabaseUrl, log.New(os.Stdout, "[MIGRATION] ", log.LstdFlags)); err != nil {
		return err
	}

	sqlDB, err := database.NewConnection(cfg.DatabaseUrl)
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	store, err := buckets.NewClient(cfg)
	if err != nil {
		return err
	}

	startupCtx, cancelStartup := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelStartup()
	if err := store.EnsureBucket(startupCtx); err != nil {
		return err
	}

	querier := database.New(sqlDB)
	emailSender := email.NewSender(cfg)
	avatarService := avatars.NewService(store)
	apiServer := handler.NewServer(querier, logger, cfg, emailSender, store, avatarService)

	r := NewRouter(apiServer, logger)

	httpServer := &http.Server{
		Addr:         ":" + cfg.HTTPPort,
		Handler:      r,
		ReadTimeout:  5 * time.Minute,
		WriteTimeout: 5 * time.Minute,
		IdleTimeout:  120 * time.Second,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go handler.StartSessionSweeper(ctx, querier, logger)

	go func() {
		logger.Info("Server starting", "port", cfg.HTTPPort)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("Server ListenAndServe error", "err", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server forced to shutdown", "err", err)
	}
	logger.Info("Server exiting.")
	return nil
}
