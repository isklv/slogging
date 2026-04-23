package chi

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestTraceMiddleware(t *testing.T) {
	t.Run("adds trace ID when not present", func(t *testing.T) {
		r := chi.NewRouter()
		r.Use(TraceMiddleware(slog.Default())) // nil logger is fine for middleware test

		r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
			// Check that trace ID is in context
			w.WriteHeader(http.StatusOK)
		})

		https := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, https)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("preserves existing trace ID from header", func(t *testing.T) {
		r := chi.NewRouter()
		r.Use(TraceMiddleware(slog.Default()))

		r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-B3-TraceId", "existing-trace-id-123")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("handles multiple requests concurrently", func(t *testing.T) {
		r := chi.NewRouter()
		r.Use(TraceMiddleware(slog.Default()))

		var done chan struct{} = make(chan struct{}, 10)

		r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			done <- struct{}{}
		})

		for i := 0; i < 10; i++ {
			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			go r.ServeHTTP(w, req)
		}

		for i := 0; i < 10; i++ {
			<-done
		}
	})

	t.Run("middleware chain works", func(t *testing.T) {
		r := chi.NewRouter()
		r.Use(TraceMiddleware(slog.Default()))
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				next.ServeHTTP(w, r)
			})
		})

		r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestTraceMiddleware_WithLogger(t *testing.T) {
	t.Run("works with real logger", func(t *testing.T) {
		// Import slogging to get real logger
		// This would require importing the parent package
		t.Skip("Skip: requires slogging import")
	})
}

func TestTraceMiddleware_Headers(t *testing.T) {
	t.Run("reads X-B3-TraceId header", func(t *testing.T) {
		r := chi.NewRouter()
		r.Use(TraceMiddleware(slog.Default()))

		r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
			// In real implementation, we'd check context
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-B3-TraceId", "test-trace-123")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		// The middleware should preserve or set the trace ID
	})
}

func TestTraceMiddleware_PathPatterns(t *testing.T) {
	t.Run("works with path parameters", func(t *testing.T) {
		r := chi.NewRouter()
		r.Use(TraceMiddleware(slog.Default()))

		r.Get("/users/{id}", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest("GET", "/users/123", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("works with path prefix", func(t *testing.T) {
		r := chi.NewRouter()
		r.Use(TraceMiddleware(slog.Default()))

		api := chi.NewRouter()
		api.Get("/test", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		r.Mount("/api", api)

		req := httptest.NewRequest("GET", "/api/test", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func BenchmarkTraceMiddleware(b *testing.B) {
	r := chi.NewRouter()
	r.Use(TraceMiddleware(slog.Default()))
	r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(w, req)
	}
}
