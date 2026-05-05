package slogging

import (
	"bytes"
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestGraylogHandlerCreation tests handler creation
func TestGraylogHandlerCreation(t *testing.T) {
	t.Run("creates handler with empty writer", func(t *testing.T) {
		h := Option{}.NewGraylogHandler()
		assert.NotNil(t, h)
	})

	t.Run("creates handler with level", func(t *testing.T) {
		h := Option{Level: slog.LevelInfo}.NewGraylogHandler()
		assert.NotNil(t, h)
	})
}

// TestGraylogHandlerEnabledMethod tests Enabled method
func TestGraylogHandlerEnabledMethod(t *testing.T) {
	h := Option{Level: slog.LevelInfo}.NewGraylogHandler()

	t.Run("returns true for levels >= threshold", func(t *testing.T) {
		assert.True(t, h.Enabled(context.Background(), slog.LevelInfo))
		assert.True(t, h.Enabled(context.Background(), slog.LevelWarn))
		assert.True(t, h.Enabled(context.Background(), slog.LevelError))
		assert.False(t, h.Enabled(context.Background(), slog.LevelDebug))
	})

	t.Run("returns true for all levels when debug", func(t *testing.T) {
		hDebug := Option{Level: slog.LevelDebug}.NewGraylogHandler()
		assert.True(t, hDebug.Enabled(context.Background(), slog.LevelDebug))
		assert.True(t, hDebug.Enabled(context.Background(), slog.LevelInfo))
	})
}

// TestGraylogHandlerHandle tests Handle method
func TestGraylogHandlerHandle(t *testing.T) {
	t.Run("handles record without writer", func(t *testing.T) {
		h := Option{}.NewGraylogHandler()

		rec := slog.Record{}
		rec.Message = "test message"
		rec.Level = slog.LevelInfo
		rec.Time = time.Now()

		err := h.Handle(context.Background(), rec)
		assert.NoError(t, err)
	})

	t.Run("handles record with attributes", func(t *testing.T) {
		h := Option{}.NewGraylogHandler()

		rec := slog.Record{}
		rec.Message = "test with attrs"
		rec.Level = slog.LevelInfo
		rec.AddAttrs(
			slog.String("key1", "value1"),
			slog.Int("key2", 123),
		)

		err := h.Handle(context.Background(), rec)
		assert.NoError(t, err)
	})

	t.Run("handles different log levels", func(t *testing.T) {
		levels := []slog.Level{slog.LevelDebug, slog.LevelInfo, slog.LevelWarn, slog.LevelError}

		for _, level := range levels {
			t.Run(level.String(), func(t *testing.T) {
				h := Option{}.NewGraylogHandler()

				rec := slog.Record{}
				rec.Message = "test"
				rec.Level = level

				err := h.Handle(context.Background(), rec)
				assert.NoError(t, err)
			})
		}
	})

	t.Run("handles empty record", func(t *testing.T) {
		h := Option{}.NewGraylogHandler()
		rec := slog.Record{}
		err := h.Handle(context.Background(), rec)
		assert.NoError(t, err)
	})

	t.Run("handles very long message", func(t *testing.T) {
		h := Option{}.NewGraylogHandler()
		rec := slog.Record{}
		rec.Message = string(bytes.Repeat([]byte("x"), 10000))
		rec.Level = slog.LevelInfo
		err := h.Handle(context.Background(), rec)
		assert.NoError(t, err)
	})

	t.Run("handles unicode message", func(t *testing.T) {
		h := Option{}.NewGraylogHandler()
		rec := slog.Record{}
		rec.Message = "тест сообщение 🚀"
		rec.Level = slog.LevelInfo
		err := h.Handle(context.Background(), rec)
		assert.NoError(t, err)
	})

	t.Run("handles nil context", func(t *testing.T) {
		h := Option{}.NewGraylogHandler()
		rec := slog.Record{}
		rec.Message = "test"
		assert.NotPanics(t, func() {
			h.Handle(nil, rec)
		})
	})
}

// TestGraylogHandlerWithAttrs tests WithAttrs method
func TestGraylogHandlerWithAttrs(t *testing.T) {
	t.Run("returns new handler with attributes", func(t *testing.T) {
		h := Option{}.NewGraylogHandler()
		newH := h.WithAttrs([]slog.Attr{slog.String("key", "value")})
		assert.NotNil(t, newH)
		assert.NotEqual(t, h, newH)
	})

	t.Run("returns handler with nil attrs", func(t *testing.T) {
		h := Option{}.NewGraylogHandler()
		newH := h.WithAttrs(nil)
		assert.NotNil(t, newH)
	})

	t.Run("returns handler with empty attrs", func(t *testing.T) {
		h := Option{}.NewGraylogHandler()
		newH := h.WithAttrs([]slog.Attr{})
		assert.NotNil(t, newH)
	})
}

// TestGraylogHandlerWithGroup tests WithGroup method
func TestGraylogHandlerWithGroup(t *testing.T) {
	t.Run("returns new handler with group", func(t *testing.T) {
		h := Option{}.NewGraylogHandler()
		newH := h.WithGroup("mygroup")
		assert.NotNil(t, newH)
	})

	t.Run("returns same handler with empty group", func(t *testing.T) {
		h := Option{}.NewGraylogHandler()
		newH := h.WithGroup("")
		assert.Equal(t, h, newH)
	})

	t.Run("nested groups", func(t *testing.T) {
		h := Option{}.NewGraylogHandler()
		h1 := h.WithGroup("g1")
		h2 := h1.WithGroup("g2")
		assert.NotNil(t, h2)
		assert.NotEqual(t, h, h1)
		assert.NotEqual(t, h1, h2)
	})
}

// TestGraylogHandlerConcurrency tests concurrent safety
func TestGraylogHandlerConcurrency(t *testing.T) {
	h := Option{}.NewGraylogHandler()

	done := make(chan bool, 100)

	for i := 0; i < 100; i++ {
		go func(id int) {
			rec := slog.Record{}
			rec.Message = "concurrent message"
			rec.Level = slog.LevelInfo
			rec.AddAttrs(slog.Int("id", id))

			h.Handle(context.Background(), rec)
			done <- true
		}(i)
	}

	for i := 0; i < 100; i++ {
		<-done
	}
}

// TestGraylogHandlerEdgeCases tests edge cases
func TestGraylogHandlerEdgeCases(t *testing.T) {
	t.Run("message with newlines", func(t *testing.T) {
		h := Option{}.NewGraylogHandler()
		rec := slog.Record{}
		rec.Message = "line1\nline2\nline3"
		rec.Level = slog.LevelInfo
		err := h.Handle(context.Background(), rec)
		assert.NoError(t, err)
	})

	t.Run("message with special chars", func(t *testing.T) {
		h := Option{}.NewGraylogHandler()
		rec := slog.Record{}
		rec.Message = "test \"quoted\" and 'single'"
		rec.Level = slog.LevelInfo
		err := h.Handle(context.Background(), rec)
		assert.NoError(t, err)
	})

	t.Run("empty message", func(t *testing.T) {
		h := Option{}.NewGraylogHandler()
		rec := slog.Record{}
		rec.Message = ""
		rec.Level = slog.LevelInfo
		err := h.Handle(context.Background(), rec)
		assert.NoError(t, err)
	})

	t.Run("whitespace only message", func(t *testing.T) {
		h := Option{}.NewGraylogHandler()
		rec := slog.Record{}
		rec.Message = "   "
		rec.Level = slog.LevelInfo
		err := h.Handle(context.Background(), rec)
		assert.NoError(t, err)
	})
}

// TestGraylogHandlerMethods tests all handler interface methods
func TestGraylogHandlerMethods(t *testing.T) {
	h := Option{}.NewGraylogHandler()

	// Test Enabled
	t.Run("Enabled method", func(t *testing.T) {
		enabled := h.Enabled(context.Background(), slog.LevelInfo)
		assert.True(t, enabled)
	})

	// Test Handle
	t.Run("Handle method", func(t *testing.T) {
		rec := slog.Record{}
		rec.Message = "test"
		err := h.Handle(context.Background(), rec)
		assert.NoError(t, err)
	})

	// Test WithAttrs
	t.Run("WithAttrs method", func(t *testing.T) {
		newH := h.WithAttrs([]slog.Attr{slog.String("key", "value")})
		assert.NotNil(t, newH)
		// Verify it implements slog.Handler
		var _ slog.Handler = newH
	})

	// Test WithGroup
	t.Run("WithGroup method", func(t *testing.T) {
		newH := h.WithGroup("group")
		assert.NotNil(t, newH)
		var _ slog.Handler = newH
	})
}

// TestGraylogHandlerLevels tests all log levels including custom ones
func TestGraylogHandlerLevels(t *testing.T) {
	levels := []struct {
		name  string
		level slog.Level
	}{
		{"Debug", slog.LevelDebug},
		{"Info", slog.LevelInfo},
		{"Warn", slog.LevelWarn},
		{"Error", slog.LevelError},
		{"Fatal", LevelFatal},
	}

	for _, tc := range levels {
		t.Run(tc.name, func(t *testing.T) {
			h := Option{}.NewGraylogHandler()
			rec := slog.Record{}
			rec.Message = tc.name
			rec.Level = tc.level

			err := h.Handle(context.Background(), rec)
			assert.NoError(t, err)
		})
	}
}

// TestGraylogHandlerWithSource tests AddSource option
func TestGraylogHandlerWithSource(t *testing.T) {
	h := Option{AddSource: true}.NewGraylogHandler()
	rec := slog.Record{}
	rec.Message = "test"
	rec.Level = slog.LevelInfo
	err := h.Handle(context.Background(), rec)
	assert.NoError(t, err)
}

// TestGraylogHandlerWithReplaceAttr tests ReplaceAttr option
func TestGraylogHandlerWithReplaceAttr(t *testing.T) {
	replaceFunc := func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == "password" {
			return slog.Attr{} // Remove password
		}
		return a
	}

	h := Option{ReplaceAttr: replaceFunc}.NewGraylogHandler()
	rec := slog.Record{}
	rec.Message = "test"
	rec.Level = slog.LevelInfo
	err := h.Handle(context.Background(), rec)
	assert.NoError(t, err)
}

// TestGraylogHandlerWithAttrFromContext tests AttrFromContext option
func TestGraylogHandlerWithAttrFromContext(t *testing.T) {
	extractor := func(ctx context.Context) []slog.Attr {
		return []slog.Attr{slog.String("trace", "abc123")}
	}

	h := Option{AttrFromContext: []func(context.Context) []slog.Attr{extractor}}.NewGraylogHandler()
	ctx := context.Background()

	rec := slog.Record{}
	rec.Message = "test"
	rec.Level = slog.LevelInfo
	err := h.Handle(ctx, rec)
	assert.NoError(t, err)
}

// TestGraylogHandlerNilWriter tests nil writer handling
func TestGraylogHandlerNilWriter(t *testing.T) {
	h := Option{Writer: nil}.NewGraylogHandler()
	rec := slog.Record{}
	rec.Message = "test"
	err := h.Handle(context.Background(), rec)
	assert.NoError(t, err)
}

// TestGraylogHandlerShort tests short function
func TestGraylogHandlerShort(t *testing.T) {
	t.Run("truncates at newline", func(t *testing.T) {
		rec := &slog.Record{}
		rec.Message = "line1\nline2"
		result := short(rec)
		assert.Equal(t, "line1", result)
	})

	t.Run("returns full message without newline", func(t *testing.T) {
		rec := &slog.Record{}
		rec.Message = "no newline here"
		result := short(rec)
		assert.Equal(t, "no newline here", result)
	})

	t.Run("trims whitespace", func(t *testing.T) {
		rec := &slog.Record{}
		rec.Message = "  trimmed  "
		result := short(rec)
		assert.Equal(t, "trimmed", result)
	})

	t.Run("handles empty", func(t *testing.T) {
		rec := &slog.Record{}
		rec.Message = ""
		result := short(rec)
		assert.Equal(t, "", result)
	})
}

// TestGraylogLogLevels tests LogLevels map
func TestGraylogLogLevels(t *testing.T) {
	t.Run("contains debug", func(t *testing.T) {
		assert.NotNil(t, LogLevels[slog.LevelDebug])
		assert.Equal(t, int32(7), LogLevels[slog.LevelDebug])
	})

	t.Run("contains info", func(t *testing.T) {
		assert.NotNil(t, LogLevels[slog.LevelInfo])
		assert.Equal(t, int32(6), LogLevels[slog.LevelInfo])
	})

	t.Run("contains warn", func(t *testing.T) {
		assert.NotNil(t, LogLevels[slog.LevelWarn])
		assert.Equal(t, int32(4), LogLevels[slog.LevelWarn])
	})

	t.Run("contains error", func(t *testing.T) {
		assert.NotNil(t, LogLevels[slog.LevelError])
		assert.Equal(t, int32(3), LogLevels[slog.LevelError])
	})

	t.Run("contains fatal", func(t *testing.T) {
		assert.NotNil(t, LogLevels[LevelFatal])
		assert.Equal(t, int32(2), LogLevels[LevelFatal])
	})
}

// TestGraylogHandlerB3TraceID tests XB3TraceID constant
func TestGraylogHandlerB3TraceID(t *testing.T) {
	assert.Equal(t, "X-B3-TraceId", XB3TraceID)
}
