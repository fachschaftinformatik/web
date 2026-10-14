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
	"github.com/fachschaftinformatik/web/internal/config"
	"github.com/fachschaftinformatik/web/internal/database"
	"github.com/fachschaftinformatik/web/internal/email"
	"github.com/fachschaftinformatik/web/internal/storage/s3"
)

var Version = "dev"

func Run() error {
	cfg := config.New()

	// Initialize structured logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
		ReplaceAttr: func(groups []string, attr slog.Attr) slog.Attr {
			if len(groups) == 0 {
				switch attr.Key {
				case slog.MessageKey:
					attr.Key = "message"
				case slog.TimeKey:
					attr.Value = slog.StringValue(attr.Value.Time().Format(time.RFC3339))
				}
			}
			return attr
		},
	})).With("service", "fachschaft-api")

	if err := database.Migrate(cfg.DatabaseUrl, log.New(os.Stdout, "[MIGRATION] ", log.LstdFlags)); err != nil {
		return err
	}

	sqlDB, err := database.NewConnection(cfg.DatabaseUrl)
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	store, err := s3.NewClient(cfg)
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

	go func() {
		logger.Info("Server starting", "port", cfg.HTTPPort, "version", Version)
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
