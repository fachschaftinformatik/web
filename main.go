package main

//go:generate go tool swag init

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
	"github.com/fachschaftinformatik/web/internal/avatars"
	"github.com/fachschaftinformatik/web/internal/buckets"
	"github.com/fachschaftinformatik/web/internal/config"
	"github.com/fachschaftinformatik/web/internal/database"
	"github.com/fachschaftinformatik/web/internal/email"
	"github.com/fachschaftinformatik/web/internal/middleware"

	_ "github.com/fachschaftinformatik/web/docs"
	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger/v2"

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
	avatarService := avatars.NewService(store)
	authServer := auth.NewServer(querier, logger, cfg, emailSender, store, avatarService)

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
		r.Put("/auth/me", authServer.PutAuthMe)
		r.Post("/auth/register", authServer.PostAuthRegister)
		r.Get("/auth/verify", authServer.GetAuthVerify)
		r.Get("/auth/notifications", authServer.GetAuthNotifications)
		r.Put("/auth/notifications/{id}/read", authServer.PutAuthNotificationsIdRead)
		r.Put("/auth/notifications/read-all", authServer.PutAuthNotificationsReadAll)
		r.Get("/auth/avatars/{userId}/{filename}", authServer.GetAvatar)

		r.Get("/users", authServer.GetUsers)
		r.Get("/users/{id}", authServer.GetUsersId)

		r.Get("/programs", authServer.GetPrograms)
		r.Get("/programs/{id}", authServer.GetProgramsId)
		r.Get("/programs/{id}/modules", authServer.GetProgramModules)

		r.Get("/exams", authServer.GetExams)
		r.Post("/exams", authServer.PostExams)
		r.Get("/exams/{id}", authServer.GetExamsId)
		r.Put("/exams/{id}", authServer.PutExamsId)
		r.Delete("/exams/{id}", authServer.DeleteExamsId)
		r.Get("/exams/{id}/file", authServer.GetExamsFile)
		r.Get("/exams/versions/{groupId}", authServer.GetExamVersions)
		r.Get("/search", authServer.GetSearch)

		r.Get("/events", authServer.GetEvents)
		r.Post("/events", authServer.PostEvents)
		r.Get("/events/{id}/cover", authServer.GetEventCover)
		r.Get("/media", authServer.GetMedia)
		r.Post("/media", authServer.PostMedia)
		r.Get("/media/{id}", authServer.GetMediaById)
		r.Get("/media/{id}/file", authServer.GetMediaFile)

		// Forum
		r.Get("/forum/posts", authServer.GetForumPosts)
		r.Post("/forum/posts", authServer.PostForumPosts)
		r.Get("/forum/posts/{id}", authServer.GetForumPostsId)
		r.Put("/forum/posts/{id}", authServer.PutForumPostsId)
		r.Delete("/forum/posts/{id}", authServer.DeleteForumPostsId)
		r.Get("/forum/posts/{id}/comments", authServer.GetForumPostsComments)
		r.Post("/forum/posts/{id}/comments", authServer.PostForumPostsComments)
		r.Put("/forum/comments/{id}", authServer.PutForumPostsCommentsId) // Update comment route
		r.Post("/forum/comments/{id}/vote", authServer.PostForumCommentsVote)
		r.Post("/forum/posts/{id}/vote", authServer.PostForumPostsVote)

		// Activities
		r.Get("/activities", authServer.GetActivities)
		r.Get("/users/{id}/activities", authServer.GetUsersActivities)

	})

	httpServer := &http.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: r,
		// Erhöht auf 5 Minuten für große Uploads
		ReadTimeout:  5 * time.Minute,
		WriteTimeout: 5 * time.Minute,
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
