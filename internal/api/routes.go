package api

import (
	"net/http"
	"time"

	"github.com/fachschaftinformatik/web/internal/api/handler"
	"github.com/fachschaftinformatik/web/internal/api/middleware"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httplog/v2"
	"github.com/go-chi/httprate"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func NewRouter(s *handler.Server, logger *httplog.Logger) http.Handler {
	r := chi.NewRouter()
	r.Use(chiMiddleware.RealIP)
	r.Use(chiMiddleware.RequestID)
	r.Use(httplog.Handler(logger))
	r.Use(middleware.SecurityHeaders)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{s.Config.Domain},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link", "X-Total-Count"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	r.Use(chiMiddleware.Recoverer)
	r.Use(httprate.LimitByIP(100, 1*time.Minute))

	r.Get("/swagger/*", httpSwagger.WrapHandler)

	r.Route("/api", func(r chi.Router) {
		r.Get("/auth/csrf", s.GetAuthCsrf)

		r.Group(func(r chi.Router) {
			r.Use(httprate.LimitByIP(20, 1*time.Minute))
			r.Post("/auth/login", s.PostAuthLogin)
			r.Post("/auth/register", s.PostAuthRegister)
		})

		r.Get("/auth/verify", s.GetAuthVerify)
		r.Get("/auth/avatars/{userId}/{filename}", s.GetAvatar)

		r.Get("/programs", s.GetPrograms)
		r.Get("/programs/{id}", s.GetProgramsId)
		r.Get("/programs/{id}/modules", s.GetProgramModules)

		r.Get("/events", s.GetEvents)
		r.Get("/events/{id}/cover", s.GetEventCover)
		r.Get("/media", s.GetMedia)
		r.Get("/media/{id}", s.GetMediaById)
		r.Get("/media/{id}/file", s.GetMediaFile)
		r.Get("/media/{id}/preview", s.GetMediaPreview)

		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireCSRF(s))
			r.Use(middleware.OptionalAuth(s))
			r.Get("/forum/posts", s.GetForumPosts)
			r.Get("/forum/posts/{id}", s.GetForumPostsId)
			r.Get("/forum/posts/{id}/comments", s.GetForumPostsComments)
			r.Get("/users/{id}", s.GetUsersId)
			r.Get("/users/{id}/activities", s.GetUsersActivities)
			r.Get("/search", s.GetSearch)
		})

		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireCSRF(s))
			r.Use(middleware.RequireAuth(s))

			r.Post("/auth/logout", s.PostAuthLogout)
			r.Get("/auth/me", s.GetAuthMe)
			r.Put("/auth/me", s.PutAuthMe)
			r.Get("/auth/notifications", s.GetAuthNotifications)
			r.Put("/auth/notifications/{id}/read", s.PutAuthNotificationsIdRead)
			r.Put("/auth/notifications/read-all", s.PutAuthNotificationsReadAll)

			r.Get("/exams", s.GetExams)
			r.Get("/exams/{id}", s.GetExamsId)
			r.Get("/exams/{id}/file", s.GetExamsFile)
			r.Get("/exams/versions/{groupId}", s.GetExamVersions)

			r.Get("/activities", s.GetActivities)

			r.Post("/forum/posts", s.PostForumPosts)
			r.Put("/forum/posts/{id}", s.PutForumPostsId)
			r.Delete("/forum/posts/{id}", s.DeleteForumPostsId)
			r.Post("/forum/posts/{id}/comments", s.PostForumPostsComments)
			r.Put("/forum/comments/{id}", s.PutForumPostsCommentsId)
			r.Post("/forum/comments/{id}/vote", s.PostForumCommentsVote)
			r.Post("/forum/posts/{id}/vote", s.PostForumPostsVote)

			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole("editor", "admin"))
				r.Post("/exams", s.PostExams)
				r.Put("/exams/{id}", s.PutExamsId)
				r.Delete("/exams/{id}", s.DeleteExamsId)
				r.Post("/events", s.PostEvents)
				r.Post("/media", s.PostMedia)
			})

			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole("admin"))
				r.Get("/users", s.GetUsers)
			})
		})
	})

	return r
}
