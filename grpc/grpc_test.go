package grpc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestTraceInterceptor(t *testing.T) {
	t.Run("returns unary interceptor", func(t *testing.T) {
		interceptor := TraceInterceptor(nil)
		assert.NotNil(t, interceptor)
	})

	t.Run("interceptor signature", func(t *testing.T) {
		interceptor := TraceInterceptor(nil)
		
		// Test that it has correct signature by calling it
		var ctx context.Context
		var req interface{}
		var info *grpc.UnaryServerInfo
		var handler grpc.UnaryHandler

		// This should not panic
		assert.NotPanics(t, func() {
			_, _ = interceptor(ctx, req, info, handler)
		})
	})
}

func TestTraceMetadata(t *testing.T) {
	t.Run("adds metadata to context", func(t *testing.T) {
		ctx := context.Background()
		newCtx := TraceMetadata(ctx)
		
		assert.NotNil(t, newCtx)
		assert.NotSame(t, ctx, newCtx)
	})

	t.Run("preserves existing context values", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "key", "value")
		newCtx := TraceMetadata(ctx)
		
		assert.Equal(t, "value", newCtx.Value("key"))
	})

	t.Run("handles nil context", func(t *testing.T) {
		// Should handle gracefully or use background
		assert.NotPanics(t, func() {
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
		interceptor := TraceInterceptor(nil)
		
		info := &grpc.UnaryServerInfo{
			FullMethod: "/package.Service/Method",
		}
		
		assert.NotPanics(t, func() {
			_, _ = interceptor(context.Background(), nil, info, nil)
		})
	})
}

func TestTraceInterceptor_Handler(t *testing.T) {
	t.Run("wraps handler execution", func(t *testing.T) {
		called := false
		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			called = true
			return nil, nil
		}
		
		interceptor := TraceInterceptor(nil)
		
		_, _ = interceptor(context.Background(), nil, nil, handler)
		
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

func BenchmarkTraceInterceptor(b *testing.B) {
	interceptor := TraceInterceptor(nil)
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, nil
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = interceptor(context.Background(), nil, nil, handler)
	}
}

func BenchmarkTraceMetadata(b *testing.B) {
	ctx := context.Background()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = TraceMetadata(ctx)
	}
}
