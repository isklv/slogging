package gin

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestTraceMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("adds trace ID when not present", func(t *testing.T) {
		r := gin.New()
		r.Use(TraceMiddleware(slog.Default()))

		r.GET("/test", func(c *gin.Context) {
			c.String(http.StatusOK, "ok")
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "ok", w.Body.String())
	})

	t.Run("preserves existing trace ID", func(t *testing.T) {
		r := gin.New()
		r.Use(TraceMiddleware(slog.Default()))

		r.GET("/test", func(c *gin.Context) {
			traceID := c.Request.Header.Get("X-B3-TraceId")
			c.String(http.StatusOK, traceID)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-B3-TraceId", "existing-trace-456")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "existing-trace-456", w.Body.String())
	})

	t.Run("handles concurrent requests", func(t *testing.T) {
		r := gin.New()
		r.Use(TraceMiddleware(slog.Default()))

		done := make(chan bool, 10)

		r.GET("/test", func(c *gin.Context) {
			c.String(http.StatusOK, "ok")
			done <- true
		})

		for i := 0; i < 10; i++ {
			go func() {
				req := httptest.NewRequest("GET", "/test", nil)
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)
			}()
		}

		for i := 0; i < 10; i++ {
			<-done
		}
	})

	t.Run("works with groups", func(t *testing.T) {
		r := gin.New()
		api := r.Group("/api")
		api.Use(TraceMiddleware(slog.Default()))

		api.GET("/users", func(c *gin.Context) {
			c.String(http.StatusOK, "users")
		})

		req := httptest.NewRequest("GET", "/api/users", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("middleware chain", func(t *testing.T) {
		r := gin.New()
		r.Use(TraceMiddleware(slog.Default()))
		r.Use(func(c *gin.Context) {
			c.Next()
		})

		r.GET("/test", func(c *gin.Context) {
			c.String(http.StatusOK, "ok")
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestTraceMiddleware_Headers(t *testing.T) {
	t.Run("passes headers through", func(t *testing.T) {
		r := gin.New()
		r.Use(TraceMiddleware(slog.Default()))

		r.GET("/test", func(c *gin.Context) {
			contentType := c.Request.Header.Get("Content-Type")
			c.String(http.StatusOK, contentType)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Body.String())
	})
}

func BenchmarkTraceMiddleware(b *testing.B) {
	r := gin.New()
	r.Use(TraceMiddleware(slog.Default()))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(w, req)
	}
}
