package grpc

import (
	"context"
	"github.com/isklv/slogging"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"log/slog"
)

func TraceInterceptor(l *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			traceID := slogging.GenerateTraceID()
			ctx = slogging.ContextWithLogger(ctx, &slogging.SLogger{l.With(slogging.StringAttr(slogging.XB3TraceID, traceID))})
			ctx = context.WithValue(ctx, slogging.XB3TraceID, traceID)
			return handler(ctx, req)
		}

		traceIDs := md.Get(slogging.XB3TraceID)

		var traceID string
		if len(traceIDs) > 0 {
			traceID = traceIDs[0]
		} else {
			traceID = slogging.GenerateTraceID()
		}

		ctx = slogging.ContextWithLogger(ctx, &slogging.SLogger{l.With(slogging.StringAttr(slogging.XB3TraceID, traceID))})
		ctx = context.WithValue(ctx, slogging.XB3TraceID, traceID)
		return handler(ctx, req)
	}
}

func TraceMetadata(ctx context.Context) context.Context {
	traceID, ok := ctx.Value(slogging.XB3TraceID).(string)
	if !ok || traceID == "" {
		traceID = slogging.GenerateTraceID()
	}

	md := metadata.Pairs(slogging.XB3TraceID, traceID)
	return metadata.NewOutgoingContext(ctx, md)
}
