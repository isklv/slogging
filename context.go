package slogging

import (
	"context"
	"github.com/google/uuid"
	"log/slog"
)

type ctxLogger struct{}

func ContextWithLogger(ctx context.Context, l *SLogger) context.Context {
	return context.WithValue(ctx, ctxLogger{}, l)
}

func GenerateTraceID() string {
	traceID := uuid.New()
	return traceID.String()
}

func Context() context.Context {
	traceID := GenerateTraceID()

	l := &SLogger{slog.Default().With(StringAttr(XB3TraceID, traceID))}
	ctx := context.WithValue(context.Background(), XB3TraceID, traceID)
	return context.WithValue(ctx, ctxLogger{}, l)
}

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
