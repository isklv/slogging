package http

import (
	"context"
	"github.com/isklv/slogging"
	"log/slog"
	"net/http"
)

func TraceRequest(ctx context.Context, req *http.Request) *http.Request {
	traceID, ok := ctx.Value(slogging.XB3TraceID).(string)
	if !ok || traceID == "" {
		traceID = slogging.GenerateTraceID()
	}

	req.Header.Set(slogging.XB3TraceID, traceID)
	return req
}

func TraceMiddleware(l *slog.Logger) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			traceID := r.Header.Get(slogging.XB3TraceID)
			if traceID == "" {
				traceID = slogging.GenerateTraceID()
			}

			newL := l.With(slogging.StringAttr(slogging.XB3TraceID, traceID))
			sl := &slogging.SLogger{Logger: newL}

			ctx := slogging.ContextWithLogger(r.Context(), sl)
			ctx = context.WithValue(ctx, slogging.XB3TraceID, traceID)

			w.Header().Set(slogging.XB3TraceID, traceID)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	}
}
