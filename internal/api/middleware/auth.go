package middleware

import (
	"context"
	"net/http"

	"github.com/fachschaftinformatik/web/internal/database"
)

type AuthProvider interface {
	Authenticate(w http.ResponseWriter, r *http.Request) (database.Session, database.User, error)
	JsonError(w http.ResponseWriter, err, msg string, status int)
}

const (
	UserKey    = "user"
	SessionKey = "session"
)

func RequireAuth(provider AuthProvider) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, user, err := provider.Authenticate(w, r)
			if err != nil {
				provider.JsonError(w, "unauthorized", "Authentication required", http.StatusUnauthorized)
				return
			}
			if user.Active == 0 {
				provider.JsonError(w, "account_disabled", "Your account has been disabled", http.StatusForbidden)
				return
			}
			if user.Verified == 0 {
				provider.JsonError(w, "email_not_verified", "Please verify your email address", http.StatusForbidden)
				return
			}
			ctx := context.WithValue(r.Context(), UserKey, user)
			ctx = context.WithValue(ctx, SessionKey, session)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func OptionalAuth(provider AuthProvider) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, user, err := provider.Authenticate(w, r)
			if err == nil && user.Active == 1 && user.Verified == 1 {
				ctx := context.WithValue(r.Context(), UserKey, user)
				ctx = context.WithValue(ctx, SessionKey, session)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := r.Context().Value(UserKey).(database.User)
			if !ok {
				// This middleware should be used after RequireAuth
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			allowed := false
			for _, role := range roles {
				if user.Role == role {
					allowed = true
					break
				}
			}
			if !allowed {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
