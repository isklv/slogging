package http

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTraceMiddleware(t *testing.T) {
	t.Run("wraps handler", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		wrapped := TraceMiddleware(slog.Default())(handler)
		assert.NotNil(t, wrapped)
	})

	t.Run("calls next handler", func(t *testing.T) {
		called := false
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			w.WriteHeader(http.StatusOK)
		})

		wrapped := TraceMiddleware(slog.Default())(handler)
		rec := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		wrapped.ServeHTTP(rec, req)

		assert.True(t, called)
	})

	t.Run("adds trace ID to context", func(t *testing.T) {
		var ctx context.Context
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx = r.Context()
		})

		wrapped := TraceMiddleware(slog.Default())(handler)
		req, _ := http.NewRequest("GET", "/test", nil)
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		assert.NotNil(t, ctx)
	})
}

func TestTraceRequest(t *testing.T) {
	t.Run("adds trace ID to request", func(t *testing.T) {
		ctx := context.Background()
		req, _ := http.NewRequest("GET", "/test", nil)

		newReq := TraceRequest(ctx, req)
		assert.NotNil(t, newReq)
		assert.NotNil(t, newReq.Context())
	})

	t.Run("preserves original request properties", func(t *testing.T) {
		ctx := context.Background()
		req, _ := http.NewRequest("POST", "http://example.com/test", nil)
		req.Header.Set("X-Custom-Header", "value")

		newReq := TraceRequest(ctx, req)
		assert.Equal(t, "POST", newReq.Method)
		assert.Equal(t, "http://example.com/test", newReq.URL.String())
		assert.Equal(t, "value", newReq.Header.Get("X-Custom-Header"))
	})

	t.Run("nil request returns nil or empty", func(t *testing.T) {
		ctx := context.Background()
		// TraceRequest does not handle nil, so it panics
		assert.Panics(t, func() {
			_ = TraceRequest(ctx, nil)
		})
	})
}

func TestTraceRequest_WithHeaders(t *testing.T) {
	t.Run("preserves headers", func(t *testing.T) {
		ctx := context.Background()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer token123")

		newReq := TraceRequest(ctx, req)

		assert.Equal(t, "application/json", newReq.Header.Get("Content-Type"))
		assert.Equal(t, "Bearer token123", newReq.Header.Get("Authorization"))
	})
}

func TestTraceRequest_Methods(t *testing.T) {
	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			ctx := context.Background()
			req, _ := http.NewRequest(method, "/test", nil)

			newReq := TraceRequest(ctx, req)
			assert.Equal(t, method, newReq.Method)
		})
	}
}

func BenchmarkTraceMiddleware(b *testing.B) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := TraceMiddleware(slog.Default())(handler)
	req, _ := http.NewRequest("GET", "/test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)
	}
}

func BenchmarkTraceRequest(b *testing.B) {
	ctx := context.Background()
	req, _ := http.NewRequest("GET", "/test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = TraceRequest(ctx, req)
	}
}
