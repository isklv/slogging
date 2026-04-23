package slogging

import (
	"context"
	"github.com/google/uuid"
	"log/slog"
)

// ctxLogger is a context key type for storing loggers.
type ctxLogger struct{}

// ContextWithLogger adds a logger to the context.
func ContextWithLogger(ctx context.Context, l *SLogger) context.Context {
	return context.WithValue(ctx, ctxLogger{}, l)
}

// GenerateTraceID generates a new UUID trace ID.
func GenerateTraceID() string {
	traceID := uuid.New()
	return traceID.String()
}

// Context creates a new context with a trace ID and logger.
func Context() context.Context {
	traceID := GenerateTraceID()

	l := &SLogger{slog.Default().With(StringAttr(XB3TraceID, traceID))}
	ctx := context.WithValue(context.Background(), XB3TraceID, traceID)
	return context.WithValue(ctx, ctxLogger{}, l)
}

// L retrieves logger from context, creating a new one with trace ID or default if unavailable.
func L(ctx context.Context) *SLogger {
	if l, ok := ctx.Value(ctxLogger{}).(*SLogger); ok {
		return l
	}

	traceID, ok := ctx.Value(XB3TraceID).(string)
	if ok {
		return &SLogger{
			Logger: slog.Default().With(StringAttr(XB3TraceID, traceID)),
		}
	}

	return &SLogger{
		Logger: slog.Default(),
	}
}
