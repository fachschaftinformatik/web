package middleware

import (
	"net/http"
)

type CSRFProvider interface {
	CheckCSRF(r *http.Request) error
	JsonError(w http.ResponseWriter, err, msg string, status int)
}

func RequireCSRF(provider CSRFProvider) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "GET" || r.Method == "HEAD" || r.Method == "OPTIONS" || r.Method == "TRACE" {
				next.ServeHTTP(w, r)
				return
			}
			if err := provider.CheckCSRF(r); err != nil {
				provider.JsonError(w, "invalid_csrf", err.Error(), http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if w.Header().Get("X-Frame-Options") == "" {
			w.Header().Set("X-Frame-Options", "SAMEORIGIN")
		}
		if w.Header().Get("X-Content-Type-Options") == "" {
			w.Header().Set("X-Content-Type-Options", "nosniff")
		}
		if w.Header().Get("Referrer-Policy") == "" {
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		}

		if w.Header().Get("Content-Security-Policy") == "" {
			w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; font-src 'self' data:; img-src 'self' data: blob:; frame-src 'self' blob:; worker-src 'self' blob:; object-src 'none';")
		}

		next.ServeHTTP(w, r)
	})
}
