package middleware

import (
	"net/http"
)

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
