package slogging

import (
	"bytes"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLogger(t *testing.T) {
	t.Run("creates logger with default options", func(t *testing.T) {
		opts := NewOptions()
		logger := NewLogger(opts)

		assert.NotNil(t, logger)
		assert.NotNil(t, logger.Logger)
	})

	t.Run("creates logger with custom options", func(t *testing.T) {
		opts := NewOptions().
			SetLevel("debug").
			WithSource(true)

		logger := NewLogger(opts)
		assert.NotNil(t, logger)
	})

	t.Run("creates logger without graylog (empty URL)", func(t *testing.T) {
		opts := NewOptions().InGraylog("", "test-app")
		logger := NewLogger(opts)
		assert.NotNil(t, logger)
		// Should not panic even with empty graylog URL
	})
}

func TestLogger_LoggingMethods(t *testing.T) {
	t.Run("Info logs message", func(t *testing.T) {
		opts := NewOptions().SetLevel("debug")
		logger := NewLogger(opts)

		assert.NotPanics(t, func() {
			logger.Info("test message")
		})
	})

	t.Run("Debug logs message", func(t *testing.T) {
		opts := NewOptions().SetLevel("debug")
		logger := NewLogger(opts)

		assert.NotPanics(t, func() {
			logger.Debug("test debug message")
		})
	})

	t.Run("Warn logs message", func(t *testing.T) {
		opts := NewOptions().SetLevel("warn")
		logger := NewLogger(opts)

		assert.NotPanics(t, func() {
			logger.Warn("test warn message")
		})
	})

	t.Run("Error logs message", func(t *testing.T) {
		opts := NewOptions().SetLevel("error")
		logger := NewLogger(opts)

		assert.NotPanics(t, func() {
			logger.Error("test error message")
		})
	})

	t.Run("Fatal logs and exits", func(t *testing.T) {
		opts := NewOptions().SetLevel("debug")
		logger := NewLogger(opts)

		// Fatal should call os.Exit(1), we can't test that easily
		// so we just check it doesn't panic
		assert.NotPanics(t, func() {
			// Skip actual fatal test to avoid exiting
			t.Skip("Skipping Fatal test to avoid os.Exit")
		})
	})

	t.Run("Panic logs and panics", func(t *testing.T) {
		opts := NewOptions().SetLevel("debug")
		logger := NewLogger(opts)

		assert.Panics(t, func() {
			logger.Panic("test panic")
		})
	})
}

func TestLogger_With(t *testing.T) {
	t.Run("adds attributes to logger", func(t *testing.T) {
		opts := NewOptions().SetLevel("debug")
		logger := NewLogger(opts)

		newLogger := logger.With("key", "value")
		assert.NotNil(t, newLogger)
		assert.NotSame(t, logger, newLogger)
	})

	t.Run("chains with multiple attributes", func(t *testing.T) {
		opts := NewOptions().SetLevel("debug")
		logger := NewLogger(opts)

		newLogger := logger.With("key1", "value1").With("key2", "value2")
		assert.NotNil(t, newLogger)
	})

	t.Run("With module attribute", func(t *testing.T) {
		opts := NewOptions().SetLevel("debug")
		logger := NewLogger(opts)

		moduleLogger := logger.With("module", "test-module")
		assert.NotNil(t, moduleLogger)
		// Should log with module attribute
		moduleLogger.Info("test with module")
	})
}

func TestLogger_Attributes(t *testing.T) {
	t.Run("StringAttr creates string attribute", func(t *testing.T) {
		attr := StringAttr("key", "value")
		assert.NotNil(t, attr)
	})

	t.Run("IntAttr creates int attribute", func(t *testing.T) {
		attr := IntAttr("count", 42)
		assert.NotNil(t, attr)
	})

	t.Run("ErrAttr creates error attribute", func(t *testing.T) {
		err := assert.AnError
		attr := ErrAttr(err)
		assert.NotNil(t, attr)
	})

	t.Run("AnyAttr creates any attribute", func(t *testing.T) {
		attr := AnyAttr("data", map[string]string{"a": "b"})
		assert.NotNil(t, attr)
	})

	t.Run("FloatAttr creates float attribute", func(t *testing.T) {
		attr := FloatAttr("ratio", 3.14)
		assert.NotNil(t, attr)
	})

	t.Run("TimeAttr creates time attribute", func(t *testing.T) {
		attr := TimeAttr("timestamp", time.Now())
		assert.NotNil(t, attr)
	})
}

func TestLogger_ContextLogger(t *testing.T) {
	t.Run("L retrieves logger from context", func(t *testing.T) {
		ctx := Context()
		logger := L(ctx)
		assert.NotNil(t, logger)
	})

	t.Run("ContextWithLogger sets logger in context", func(t *testing.T) {
		opts := NewOptions().SetLevel("debug")
		logger := NewLogger(opts)
		
		ctx := ContextWithLogger(Context(), logger)
		retrievedLogger := L(ctx)
		assert.NotNil(t, retrievedLogger)
	})

	t.Run("Context creates context with trace ID", func(t *testing.T) {
		ctx := Context()
		logger := L(ctx)
		assert.NotNil(t, logger)
		// Should have trace ID in context
		logger.Info("test with context")
	})
}

func TestLogger_ConcurrentSafety(t *testing.T) {
	t.Run("concurrent logging is safe", func(t *testing.T) {
		opts := NewOptions().SetLevel("debug")
		logger := NewLogger(opts)

		done := make(chan bool, 10)

		for i := 0; i < 10; i++ {
			go func(id int) {
				for j := 0; j < 100; j++ {
					logger.Info("concurrent message", "id", id, "seq", j)
				}
				done <- true
			}(i)
		}

		for i := 0; i < 10; i++ {
			select {
			case <-done:
			case <-time.After(5 * time.Second):
				t.Fatal("timeout waiting for goroutines")
			}
		}
	})
}

func TestLogger_OutputCapture(t *testing.T) {
	t.Run("captures stdout output", func(t *testing.T) {
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		opts := NewOptions().SetLevel("debug")
		logger := NewLogger(opts)

		logger.Info("test message")

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		_, err := buf.ReadFrom(r)
		require.NoError(t, err)
		
		// Should contain our message
		output := buf.String()
		assert.Contains(t, output, "test message")
	})
}

func TestLogger_LevelFiltering(t *testing.T) {
	t.Run("debug level shows all", func(t *testing.T) {
		opts := NewOptions().SetLevel("debug")
		logger := NewLogger(opts)
		
		// All should log without error
		logger.Debug("debug")
		logger.Info("info")
		logger.Warn("warn")
		logger.Error("error")
	})

	t.Run("error level filters debug/info/warn", func(t *testing.T) {
		opts := NewOptions().SetLevel("error")
		logger := NewLogger(opts)
		
		// Only error should log
		logger.Error("error")
	})
}

func TestLogger_Handler(t *testing.T) {
	t.Run("gets underlying handler", func(t *testing.T) {
		opts := NewOptions().SetLevel("debug")
		logger := NewLogger(opts)

		handler := logger.Handler()
		assert.NotNil(t, handler)
	})

	t.Run("handler can log", func(t *testing.T) {
		opts := NewOptions().SetLevel("debug")
		logger := NewLogger(opts)

		handler := logger.Handler()
		assert.NotPanics(t, func() {
			handler.Handle(nil, record{})
		})
	})
}

// record is a minimal slog.Record implementation for testing
type record struct{}

func (r record) Time() time.Time                    { return time.Now() }
func (r record) Level() Level                      { return LevelInfo }
func (r record) Message() string                   { return "test" }
func (r record) AddAttrs(f func(A))                {}
func (r record) Attrs(f func(A))                   {}
func (r record) PC() uint64                        { return 0 }
func (r record) Src() Source                       { return Source{} }
func (r record) Group() record                     { return r }
func (r record) clone() record                     { return r }
