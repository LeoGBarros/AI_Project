package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

// responseWriter wraps http.ResponseWriter to capture the status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Logging emits a structured JSON log entry for every HTTP request, including
// method, path, status code, duration, and correlation ID.
// Follows the field convention in TECHNICAL_BASE section 6.4.
func Logging(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(wrapped, r)

			logger.Info("http request",
				zap.String("http.method", r.Method),
				zap.String("http.path", r.URL.Path),
				zap.Int("http.status_code", wrapped.statusCode),
				zap.Int64("http.duration_ms", time.Since(start).Milliseconds()),
				zap.String("http.request_id", r.Header.Get("X-Request-ID")),
				zap.String("correlation_id", GetCorrelationID(r.Context())),
			)
		})
	}
}
