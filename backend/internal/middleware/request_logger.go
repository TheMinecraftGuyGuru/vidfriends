package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/vidfriends/backend/internal/logging"
)

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

func (rw *responseWriter) Status() int {
	if rw.status == 0 {
		return http.StatusOK
	}
	return rw.status
}

// RequestLogger decorates requests with structured logging metadata.
func RequestLogger(base *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			requestID := uuid.NewString()

			reqLogger := base.With(
				slog.String("request_id", requestID),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("remote_addr", r.RemoteAddr),
			)

			ctx := logging.WithLogger(r.Context(), reqLogger)
			ctx = logging.WithRequestID(ctx, requestID)

			wrapped := &responseWriter{ResponseWriter: w}

			defer func() {
				if rec := recover(); rec != nil {
					reqLogger.Error("panic recovered", "panic", rec)
					http.Error(wrapped, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				}
				reqLogger.Info("request completed",
					slog.Int("status", wrapped.Status()),
					slog.Duration("duration", time.Since(start)),
				)
			}()

			next.ServeHTTP(wrapped, r.WithContext(ctx))
		})
	}
}
