package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// Logger is a middleware that logs the start and end of each request,
// including the status code, latency, and path.
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Create a response wrapper to capture the status code
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		
		next.ServeHTTP(rw, r)
		
		slog.Info("http request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rw.status,
			"duration", time.Since(start),
			"remote_addr", r.RemoteAddr,
		)
	})
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}
