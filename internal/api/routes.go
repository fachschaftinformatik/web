package api

import (
	"net/http"
	"strings"
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
	r.Use(httprate.LimitByIP(1000, 1*time.Minute))

	r.Get("/swagger/*", httpSwagger.WrapHandler)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/auth/csrf", s.GetAuthCsrf)

		r.Group(func(r chi.Router) {
			r.Use(httprate.LimitByIP(20, 1*time.Minute))
			r.Post("/auth/login", s.PostAuthLogin)
			r.Post("/auth/register", s.PostAuthRegister)
		})

		r.Get("/auth/verify", s.GetAuthVerify)
		r.Get("/auth/avatars/{userId}/{filename}", s.GetAvatar)

		r.Get("/programs", s.GetPrograms)
		r.Get("/programs/{programId}", s.GetProgramsId)
		r.Get("/programs/{programId}/modules", s.GetProgramModules)

		r.Get("/events", s.GetEvents)
		r.Get("/events/{eventId}", s.GetEventsId)
		r.Get("/events/{eventId}/media", s.GetEventMedia)
		r.Get("/events/{eventId}/cover", s.GetEventCover)
		r.Get("/media/{mediaId}", s.GetMediaById)
		r.Get("/media/{mediaId}/file", s.GetMediaFile)
		r.Get("/media/{mediaId}/preview", s.GetMediaPreview)

		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireCSRF(s))
			r.Use(middleware.OptionalAuth(s))
			r.Get("/discussions", s.GetDiscussions)
			r.Get("/discussions/{postId}", s.GetDiscussionsId)
			r.Get("/discussions/{postId}/comments", s.GetDiscussionsComments)
			r.Get("/users/{userId}", s.GetUsersId)
			r.Get("/users/{userId}/activities", s.GetUsersActivities)
			r.Get("/search", s.GetSearch)
		})

		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireCSRF(s))
			r.Use(middleware.RequireAuth(s))

			r.Post("/auth/logout", s.PostAuthLogout)
			r.Get("/auth/me", s.GetAuthMe)
			r.Put("/auth/me", s.PutAuthMe)
			r.Post("/auth/me/avatar", s.PostAuthAvatar)
			r.Get("/auth/notifications", s.GetAuthNotifications)
			r.Put("/auth/notifications/{notificationId}/read", s.PutAuthNotificationsIdRead)
			r.Put("/auth/notifications/read-all", s.PutAuthNotificationsReadAll)

			r.Get("/archive", s.GetArchive)
			r.Get("/archive/{entryId}", s.GetArchiveId)
			r.Get("/archive/{entryId}/file", s.GetArchiveFile)
			r.Get("/archive/{entryId}/versions", s.GetArchiveVersions)

			r.Get("/activities", s.GetActivities)

			r.Post("/discussions", s.PostDiscussions)
			r.Put("/discussions/{postId}", s.PutDiscussionsId)
			r.Delete("/discussions/{postId}", s.DeleteDiscussionsId)
			r.Post("/discussions/{postId}/comments", s.PostDiscussionsComments)
			r.Put("/discussions/comments/{commentId}", s.PutDiscussionsCommentsId)
			r.Post("/discussions/comments/{commentId}/vote", s.PostDiscussionsCommentsVote)
			r.Post("/discussions/{postId}/vote", s.PostDiscussionsVote)

			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole("editor", "admin"))
				r.Post("/archive", s.PostArchive)
				r.Put("/archive/{entryId}", s.PutArchiveId)
				r.Delete("/archive/{entryId}", s.DeleteArchiveId)
				r.Delete("/archive/files/{fileId}", s.DeleteArchiveFile)
				r.Post("/events", s.PostEvents)
				r.Post("/events/{eventId}/media", s.PostEventMedia)
			})

			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireRole("admin"))
				r.Get("/users", s.GetUsers)
				r.Get("/admin/ref/roles", s.GetAdminRefRoles)
				r.Get("/admin/ref/discussion-types", s.GetAdminRefDiscussionTypes)
				r.Post("/admin/programs", s.PostAdminPrograms)
				r.Post("/admin/modules", s.PostAdminModules)
			})
		})
	})

	// Legacy redirect or 404 for old /api
	r.Route("/api", func(r chi.Router) {
		r.HandleFunc("/*", func(w http.ResponseWriter, r *http.Request) {
			newPath := "/api/v1" + strings.TrimPrefix(r.URL.Path, "/api")
			if r.URL.RawQuery != "" {
				newPath += "?" + r.URL.RawQuery
			}
			http.Redirect(w, r, newPath, http.StatusMovedPermanently)
		})
	})

	return r
}
