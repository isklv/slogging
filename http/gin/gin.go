package gin

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/isklv/slogging"
	"log/slog"
)

func TraceMiddleware(l *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader(slogging.XB3TraceID)
		if traceID == "" {
			traceID = slogging.GenerateTraceID()
		}

		newL := l.With(slogging.StringAttr(slogging.XB3TraceID, traceID))
		sl := &slogging.SLogger{Logger: newL}

		ctx := slogging.ContextWithLogger(c.Request.Context(), sl)
		ctx = context.WithValue(ctx, slogging.XB3TraceID, traceID)

		c.Writer.Header().Set(slogging.XB3TraceID, traceID)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
