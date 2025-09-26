package logging

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

// Span represents a logical unit of work tied to a request trace.
type Span struct {
	name   string
	logger *slog.Logger
	start  time.Time
}

// StartSpan derives a child span from the provided context, enriching the logger
// with tracing metadata. It returns the derived context and the span handle.
func StartSpan(ctx context.Context, name string) (context.Context, *Span) {
	if ctx == nil {
		ctx = context.Background()
	}

	logger := FromContext(ctx)

	traceID := TraceIDFromContext(ctx)
	if traceID == "" {
		traceID = uuid.NewString()
		ctx = WithTraceID(ctx, traceID)
		logger = logger.With(slog.String("trace_id", traceID))
	}

	parentSpanID := SpanIDFromContext(ctx)
	spanID := uuid.NewString()

	logger = logger.With(
		slog.String("span_id", spanID),
		slog.String("span_name", name),
	)
	if parentSpanID != "" {
		logger = logger.With(slog.String("parent_span_id", parentSpanID))
	}

	ctx = WithLogger(ctx, logger)
	ctx = WithSpanID(ctx, spanID)

	span := &Span{
		name:   name,
		logger: logger,
		start:  time.Now(),
	}

	return ctx, span
}

// End finalizes the span and emits a completion log entry.
func (s *Span) End() {
	if s == nil {
		return
	}
	s.logger.Info("span completed", slog.Duration("duration", time.Since(s.start)))
}
