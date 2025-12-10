package mux

import (
	"context"
	"github.com/isklv/slogging"
	"log/slog"
	"net/http"
)

func TraceMiddleware(l *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		})
	}
}
