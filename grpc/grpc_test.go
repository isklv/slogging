package grpc

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestTraceInterceptor(t *testing.T) {
	t.Run("returns unary interceptor", func(t *testing.T) {
		interceptor := TraceInterceptor(slog.Default())
		assert.NotNil(t, interceptor)
	})

	t.Run("interceptor signature", func(t *testing.T) {
		interceptor := TraceInterceptor(slog.Default())

		// Test that it has correct signature by calling it with proper values
		info := &grpc.UnaryServerInfo{}
		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			return &emptypb.Empty{}, nil
		}

		// This should not panic
		assert.NotPanics(t, func() {
			ctx := context.Background()
			_, _ = interceptor(ctx, nil, info, handler)
		})
	})
}

func TestTraceMetadata(t *testing.T) {
	t.Run("adds metadata to context", func(t *testing.T) {
		ctx := context.Background()
		newCtx := TraceMetadata(ctx)

		assert.NotNil(t, newCtx)
		// Verify that outgoing metadata is added
		md, ok := metadata.FromOutgoingContext(newCtx)
		assert.True(t, ok)
		assert.Contains(t, md, "X-B3-TraceId")
	})

	t.Run("preserves existing context values", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "key", "value")
		newCtx := TraceMetadata(ctx)

		assert.Equal(t, "value", newCtx.Value("key"))
	})

	t.Run("handles nil context", func(t *testing.T) {
		// Should panic when called with nil
		assert.Panics(t, func() {
			_ = TraceMetadata(nil)
		})
	})
}

func TestTraceMetadata_MultipleCalls(t *testing.T) {
	t.Run("can be called multiple times", func(t *testing.T) {
		ctx := context.Background()
		ctx1 := TraceMetadata(ctx)
		ctx2 := TraceMetadata(ctx1)

		assert.NotNil(t, ctx2)
	})
}

func TestTraceInterceptor_UnaryServerInfo(t *testing.T) {
	t.Run("handles different service methods", func(t *testing.T) {
		interceptor := TraceInterceptor(slog.Default())

		info := &grpc.UnaryServerInfo{
			FullMethod: "/package.Service/Method",
		}

		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			return &emptypb.Empty{}, nil
		}

		assert.NotPanics(t, func() {
			_, _ = interceptor(context.Background(), nil, info, handler)
		})
	})
}

func TestTraceInterceptor_Handler(t *testing.T) {
	t.Run("wraps handler execution", func(t *testing.T) {
		called := false
		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			called = true
			return &emptypb.Empty{}, nil
		}

		interceptor := TraceInterceptor(slog.Default())
		info := &grpc.UnaryServerInfo{}

		_, _ = interceptor(context.Background(), nil, info, handler)

		assert.True(t, called)
	})
}

func TestTraceMetadata_WithExistingMetadata(t *testing.T) {
	t.Run("preserves existing metadata", func(t *testing.T) {
		md := metadata.Pairs("key1", "value1")
		ctx := metadata.NewIncomingContext(context.Background(), md)

		newCtx := TraceMetadata(ctx)

		// Should still have original metadata
		incoming, ok := metadata.FromIncomingContext(newCtx)
		assert.True(t, ok)
		assert.Contains(t, incoming, "key1")
	})
}

func TestTraceInterceptor_WithTraceID(t *testing.T) {
	t.Run("extracts trace ID from metadata", func(t *testing.T) {
		md := metadata.Pairs("X-B3-TraceId", "test-trace-123")
		ctx := metadata.NewIncomingContext(context.Background(), md)

		interceptor := TraceInterceptor(slog.Default())
		info := &grpc.UnaryServerInfo{}

		var capturedTraceID string
		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			capturedTraceID = ctx.Value("X-B3-TraceId").(string)
			return &emptypb.Empty{}, nil
		}

		_, _ = interceptor(ctx, nil, info, handler)

		assert.Equal(t, "test-trace-123", capturedTraceID)
	})

	t.Run("generates trace ID when not present", func(t *testing.T) {
		ctx := context.Background()

		interceptor := TraceInterceptor(slog.Default())
		info := &grpc.UnaryServerInfo{}

		var capturedTraceID string
		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			capturedTraceID = ctx.Value("X-B3-TraceId").(string)
			return &emptypb.Empty{}, nil
		}

		_, _ = interceptor(ctx, nil, info, handler)

		assert.NotEmpty(t, capturedTraceID)
	})
}

func TestTraceInterceptor_Concurrent(t *testing.T) {
	t.Run("handles concurrent requests", func(t *testing.T) {
		interceptor := TraceInterceptor(slog.Default())
		info := &grpc.UnaryServerInfo{}

		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			return &emptypb.Empty{}, nil
		}

		// Run concurrent requests
		for i := 0; i < 100; i++ {
			go func() {
				ctx := context.Background()
				_, _ = interceptor(ctx, nil, info, handler)
			}()
		}
	})
}

func BenchmarkTraceInterceptor(b *testing.B) {
	interceptor := TraceInterceptor(slog.Default())
	info := &grpc.UnaryServerInfo{}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return &emptypb.Empty{}, nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = interceptor(context.Background(), nil, info, handler)
	}
}

func BenchmarkTraceMetadata(b *testing.B) {
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = TraceMetadata(ctx)
	}
}
