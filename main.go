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

	"github.com/fachschaftinformatik/web/internal/auth"
	"github.com/fachschaftinformatik/web/internal/buckets"
	"github.com/fachschaftinformatik/web/internal/config"
	"github.com/fachschaftinformatik/web/internal/database"
	"github.com/fachschaftinformatik/web/internal/email"
	"github.com/fachschaftinformatik/web/internal/middleware"

	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	_ "github.com/fachschaftinformatik/web/docs"

	_ "modernc.org/sqlite"
)

// @title Fachschaft Informatik API
// @version 1.0
// @description API for the website of the FSV Informatik
// @BasePath /api
func main() {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	cfg := config.New()
	if err := database.Migrate(cfg.DatabaseUrl, logger); err != nil {
		logger.Fatalf("Migrations failed: %v", err)
	}

	sqlDB, err := database.NewConnection(cfg.DatabaseUrl)
	if err != nil {
		logger.Fatalf("Database connection failed: %v", err)
	}
	defer sqlDB.Close()

	store, err := buckets.NewClient(cfg)
	if err != nil {
		logger.Fatalf("Storage client creation failed: %v", err)
	}
	
	startupCtx, cancelStartup := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelStartup()
	if err := store.EnsureBucket(startupCtx); err != nil {
		logger.Fatalf("Failed to ensure bucket exist: %v", err)
	}

	querier := database.New(sqlDB)
	emailSender := email.NewSender(cfg)
	authServer := auth.NewServer(querier, logger, cfg, emailSender, store)

	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return middleware.Logging(logger)(next)
	})

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	r.Route("/api", func(r chi.Router) {
		r.Get("/auth/csrf", authServer.GetAuthCsrf)
		r.Post("/auth/login", authServer.PostAuthLogin)
		r.Post("/auth/logout", authServer.PostAuthLogout)
		r.Get("/auth/me", authServer.GetAuthMe)
		r.Post("/auth/register", authServer.PostAuthRegister)
		r.Get("/auth/verify", authServer.GetAuthVerify)
		
		r.Get("/users", authServer.GetUsers)
		r.Get("/users/{id}", authServer.GetUsersId)

		r.Get("/programs", authServer.GetPrograms)
		r.Get("/programs/{id}", authServer.GetProgramsId)
		r.Get("/programs/{id}/modules", authServer.GetProgramModules)

		r.Get("/exams", authServer.GetExams)
		r.Post("/exams", authServer.PostExams)
		r.Put("/exams/{id}", authServer.PutExamsId)
		r.Delete("/exams/{id}", authServer.DeleteExamsId)
		r.Get("/exams/{id}/file", authServer.GetExamsFile)
	})

	httpServer := &http.Server{
		Addr:         ":" + cfg.HTTPPort,
		Handler:      r,
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

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Printf("Server forced to shutdown: %v", err)
	}
	logger.Println("Server exiting.")
}
