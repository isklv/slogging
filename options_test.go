package slogging

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewOptions(t *testing.T) {
	t.Run("creates default options", func(t *testing.T) {
		opts := NewOptions()

		assert.NotNil(t, opts)
		assert.Equal(t, defaultLevel, opts.level)
		assert.Equal(t, defaultWithSource, opts.withSource)
		assert.Equal(t, defaultSetDefault, opts.setDefault)
		assert.Nil(t, opts.inGraylog)
	})
}

func TestOptions_SetLevel(t *testing.T) {
	t.Run("sets debug level", func(t *testing.T) {
		opts := NewOptions().SetLevel("debug")
		assert.Equal(t, slog.LevelDebug, opts.level)
	})

	t.Run("sets info level", func(t *testing.T) {
		opts := NewOptions().SetLevel("info")
		assert.Equal(t, slog.LevelInfo, opts.level)
	})

	t.Run("sets warn level", func(t *testing.T) {
		opts := NewOptions().SetLevel("warn")
		assert.Equal(t, slog.LevelWarn, opts.level)
	})

	t.Run("sets error level", func(t *testing.T) {
		opts := NewOptions().SetLevel("error")
		assert.Equal(t, slog.LevelError, opts.level)
	})
}

func TestOptions_WithSource(t *testing.T) {
	t.Run("enables source attribution", func(t *testing.T) {
		opts := NewOptions().WithSource(true)
		assert.True(t, opts.withSource)
	})

	t.Run("disables source attribution", func(t *testing.T) {
		opts := NewOptions().WithSource(false)
		assert.False(t, opts.withSource)
	})
}

func TestOptions_SetDefault(t *testing.T) {
	t.Run("sets as default logger", func(t *testing.T) {
		opts := NewOptions().SetDefault(true)
		assert.True(t, opts.setDefault)
	})

	t.Run("does not set as default", func(t *testing.T) {
		opts := NewOptions().SetDefault(false)
		assert.False(t, opts.setDefault)
	})
}

func TestOptions_InGraylog(t *testing.T) {
	t.Run("empty URL - graceful degradation", func(t *testing.T) {
		// Empty URL should not panic and should allow empty inGraylog
		opts := NewOptions().InGraylog("", "test-app")
		assert.NotNil(t, opts)
		// Note: inGraylog might be nil or have nil writer for empty URL
		// This tests that we don't crash
	})

	t.Run("invalid URL - graceful degradation", func(t *testing.T) {
		// Invalid URL should not panic
		opts := NewOptions().InGraylog("not-a-valid-url", "test-app")
		assert.NotNil(t, opts)
		// This tests that we handle error gracefully instead of log.Fatal
	})

	t.Run("valid localhost URL", func(t *testing.T) {
		// This might fail if Graylog is not running, but should not panic
		// We use a timeout port to avoid actual connection
		opts := NewOptions().InGraylog("localhost:12201", "test-app")
		assert.NotNil(t, opts)
	})

	t.Run("chaining with other options", func(t *testing.T) {
		opts := NewOptions().
			InGraylog("", "test-app").
			SetLevel("debug").
			WithSource(true).
			SetDefault(true)

		assert.Equal(t, slog.LevelDebug, opts.level)
		assert.True(t, opts.withSource)
		assert.True(t, opts.setDefault)
	})
}

func TestOptions_Chaining(t *testing.T) {
	t.Run("full chain", func(t *testing.T) {
		opts := NewOptions().
			SetLevel("debug").
			WithSource(true).
			InGraylog("", "app").
			SetDefault(true)

		assert.Equal(t, slog.LevelDebug, opts.level)
		assert.True(t, opts.withSource)
		assert.True(t, opts.setDefault)
	})
}

func TestOptions_EmptyGraylogDoesNotCrash(t *testing.T) {
	t.Run("panic test", func(t *testing.T) {
		// This should not panic
		assert.NotPanics(t, func() {
			_ = NewOptions().InGraylog("", "app")
		})
	})
}
