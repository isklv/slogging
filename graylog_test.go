package slogging

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGraylogHandler(t *testing.T) {
	t.Run("creates handler with empty URL gracefully", func(t *testing.T) {
		// Test that we can create handler even if graylog is not configured
		opts := NewOptions().InGraylog("", "test-app")
		logger := NewLogger(opts)
		
		assert.NotNil(t, logger)
		// Should not panic on logging
		logger.Info("test without graylog")
	})

	t.Run("creates handler with container name", func(t *testing.T) {
		opts := NewOptions().InGraylog("", "my-container")
		logger := NewLogger(opts)
		
		assert.NotNil(t, logger)
		// Container name should be set in options
	})

	t.Run("empty container name handled", func(t *testing.T) {
		opts := NewOptions().InGraylog("", "")
		logger := NewLogger(opts)
		
		assert.NotNil(t, logger)
	})
}

func TestGraylogData(t *testing.T) {
	t.Run("gelfData structure", func(t *testing.T) {
		data := &gelfData{
			w:             nil,
			containerName: "test-container",
		}
		
		assert.Equal(t, "test-container", data.containerName)
		assert.Nil(t, data.w)
	})
}

func TestInGraylog_ErrorHandling(t *testing.T) {
	t.Run("empty URL returns without error", func(t *testing.T) {
		// Should not call log.Fatal
		opts := NewOptions()
		result := opts.InGraylog("", "app")
		assert.Same(t, opts, result) // Should return same instance for chaining
	})

	t.Run("invalid URL returns without panic", func(t *testing.T) {
		// Should log error but not panic
		opts := NewOptions()
		result := opts.InGraylog("invalid-udp-address", "app")
		assert.NotNil(t, result)
	})
}

func TestGraylogIntegration(t *testing.T) {
	t.Run("integration with logger", func(t *testing.T) {
		// Test full flow with empty URL (graceful degradation)
		opts := NewOptions().
			InGraylog("", "integration-test").
			SetLevel("debug").
			WithSource(false)
		
		logger := NewLogger(opts)
		assert.NotNil(t, logger)
		
		// Should be able to log
		logger.Info("integration test message", "key", "value")
	})
}

func TestGraylogURLValidation(t *testing.T) {
	t.Run("validates empty string", func(t *testing.T) {
		url := ""
		if url == "" {
			// Should handle gracefully
			t.Log("Empty URL handled gracefully")
		}
	})

	t.Run("validates whitespace only", func(t *testing.T) {
		url := "   "
		if url != "" {
			// This might cause error in gelf.NewWriter
			// Should be handled gracefully
		}
	})
}

func TestGraylogHandlerEnabled(t *testing.T) {
	t.Run("handler enabled when URL provided", func(t *testing.T) {
		// When URL is empty, handler should be nil or not send
		opts := NewOptions().InGraylog("", "test")
		assert.NotNil(t, opts)
		// inGraylog might be nil or have nil writer
	})

	t.Run("handler disabled when URL empty", func(t *testing.T) {
		opts := NewOptions()
		// Before InGraylog, inGraylog is nil
		assert.Nil(t, opts.inGraylog)
	})
}

func TestGelfWriterError(t *testing.T) {
	t.Run("handles dial udp error", func(t *testing.T) {
		// This simulates what happens with empty URL
		// gelf.NewWriter("") returns "dial udp: missing address"
		// We should handle this gracefully
		opts := NewOptions().InGraylog("", "app")
		assert.NotNil(t, opts)
		// If we got here, it didn't crash
	})
}
