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

	"github.com/fachschaftinformatik/web/internal/api"
	"github.com/fachschaftinformatik/web/internal/auth"
	"github.com/fachschaftinformatik/web/internal/config"
	"github.com/fachschaftinformatik/web/internal/database"
	"github.com/fachschaftinformatik/web/internal/email"
	"github.com/fachschaftinformatik/web/internal/middleware"

	_ "modernc.org/sqlite"
)

func main() {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	cfg := config.New()

	if err := database.Migrate(cfg.GooseDSN, logger); err != nil {
		logger.Fatalf("Migrations failed: %v", err)
	}

	sqlDB, err := database.NewConnection(cfg.AppDSN)
	if err != nil {
		logger.Fatalf("Database connection failed: %v", err)
	}
	defer sqlDB.Close()
	logger.Println("Database connection established.")

	querier := database.New(sqlDB)
	emailSender := email.NewSender(cfg)
	authServer := auth.NewServer(querier, logger, cfg, emailSender)
	handler := middleware.Logging(logger)(api.Handler(authServer))
	httpServer := &http.Server{
		Addr:         ":" + cfg.HTTPPort,
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go auth.StartSessionSweeper(ctx, querier, logger)
	go func() {
		logger.Printf("Server starting on port %s", cfg.HTTPPort)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatalf("Server ListenAndServe error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Println("Shutting down server...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Printf("Server forced to shutdown: %v", err)
	}
	logger.Println("Server exiting.")
}
