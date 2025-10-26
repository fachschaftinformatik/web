package middleware

import (
	"log"
	"net/http"
	"time"
)

func Logging(logger *log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			logger.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
		})
	}
}
