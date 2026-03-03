package chi

import (
	"context"
	"net/http"

	"log/slog"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/isklv/slogging"
)

func TraceMiddleware(l *slog.Logger) func(next http.Handler) http.Handler {
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

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(ww, r.WithContext(ctx))
		})
	}
}
