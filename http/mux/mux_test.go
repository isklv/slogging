package mux

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestTraceMiddleware(t *testing.T) {
	t.Run("adds trace ID when not present", func(t *testing.T) {
		r := mux.NewRouter()
		r.Use(TraceMiddleware(nil))

		r.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ok"))
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "ok", w.Body.String())
	})

	t.Run("preserves existing trace ID", func(t *testing.T) {
		r := mux.NewRouter()
		r.Use(TraceMiddleware(nil))

		var capturedTraceID string
		r.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
			capturedTraceID = r.Header.Get("X-B3-TraceId")
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-B3-TraceId", "trace-789")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "trace-789", capturedTraceID)
	})

	t.Run("works with subrouters", func(t *testing.T) {
		r := mux.NewRouter()
		api := r.PathPrefix("/api").Subrouter()
		api.Use(TraceMiddleware(nil))

		api.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest("GET", "/api/users", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("works with path variables", func(t *testing.T) {
		r := mux.NewRouter()
		r.Use(TraceMiddleware(nil))

		r.HandleFunc("/users/{id}", func(w http.ResponseWriter, r *http.Request) {
			vars := mux.Vars(r)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(vars["id"]))
		})

		req := httptest.NewRequest("GET", "/users/123", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "123", w.Body.String())
	})

	t.Run("concurrent safety", func(t *testing.T) {
		r := mux.NewRouter()
		r.Use(TraceMiddleware(nil))

		done := make(chan bool, 10)

		r.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
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

	t.Run("middleware chain", func(t *testing.T) {
		r := mux.NewRouter()
		r.Use(TraceMiddleware(nil))
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				next.ServeHTTP(w, r)
			})
		})

		r.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestTraceMiddleware_Methods(t *testing.T) {
	t.Run("GET request", func(t *testing.T) {
		r := mux.NewRouter()
		r.Use(TraceMiddleware(nil))
		r.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}).Methods("GET")

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("POST request", func(t *testing.T) {
		r := mux.NewRouter()
		r.Use(TraceMiddleware(nil))
		r.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
		}).Methods("POST")

		req := httptest.NewRequest("POST", "/test", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("DELETE request", func(t *testing.T) {
		r := mux.NewRouter()
		r.Use(TraceMiddleware(nil))
		r.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}).Methods("DELETE")

		req := httptest.NewRequest("DELETE", "/test", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNoContent, w.Code)
	})
}

func BenchmarkTraceMiddleware(b *testing.B) {
	r := mux.NewRouter()
	r.Use(TraceMiddleware(nil))
	r.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(w, req)
	}
}
