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
	traceIDKey   ctxKey = "traceID"
	spanIDKey    ctxKey = "spanID"
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

// WithTraceID stores a trace identifier on the context.
func WithTraceID(ctx context.Context, traceID string) context.Context {
	if ctx == nil || traceID == "" {
		return ctx
	}
	return context.WithValue(ctx, traceIDKey, traceID)
}

// TraceIDFromContext retrieves the trace identifier from the context.
func TraceIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if traceID, ok := ctx.Value(traceIDKey).(string); ok {
		return traceID
	}
	return ""
}

// WithSpanID stores the current span identifier on the context.
func WithSpanID(ctx context.Context, spanID string) context.Context {
	if ctx == nil || spanID == "" {
		return ctx
	}
	return context.WithValue(ctx, spanIDKey, spanID)
}

// SpanIDFromContext retrieves the span identifier from the context.
func SpanIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if spanID, ok := ctx.Value(spanIDKey).(string); ok {
		return spanID
	}
	return ""
}
