package logging

import (
	"context"
	"log/slog"
)

// ctxKey is an unexported type for context keys defined in this package.
type ctxKey string

const (
	loggerKey    ctxKey = "logger"
	requestIDKey ctxKey = "requestID"
)

// WithLogger stores the provided logger on the context.
func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	if ctx == nil || logger == nil {
		return ctx
	}
	return context.WithValue(ctx, loggerKey, logger)
}

// FromContext returns the request-scoped logger or falls back to slog.Default().
func FromContext(ctx context.Context) *slog.Logger {
	if ctx == nil {
		return slog.Default()
	}
	if logger, ok := ctx.Value(loggerKey).(*slog.Logger); ok && logger != nil {
		return logger
	}
	return slog.Default()
}

// WithRequestID stores a request identifier on the context.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	if ctx == nil || requestID == "" {
		return ctx
	}
	return context.WithValue(ctx, requestIDKey, requestID)
}

// RequestIDFromContext retrieves a previously stored request identifier.
func RequestIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if requestID, ok := ctx.Value(requestIDKey).(string); ok {
		return requestID
	}
	return ""
}
