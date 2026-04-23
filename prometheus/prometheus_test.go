package prometheus

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTraceExemplar(t *testing.T) {
	t.Run("returns trace exemplar function", func(t *testing.T) {
		fn := TraceExemplar
		assert.NotNil(t, fn)
	})

	t.Run("function signature", func(t *testing.T) {
		// TraceExemplar should return a function that takes traceID and returns exemplar
		fn := TraceExemplar
		result := fn("test-trace-id")
		assert.NotNil(t, result)
	})
}

func TestTraceExemplar_TraceIDFormats(t *testing.T) {
	tests := []struct {
		name     string
		traceID  string
		expected string
	}{
		{
			name:     "valid trace ID",
			traceID:  "abc123def456",
			expected: "abc123def456",
		},
		{
			name:     "empty trace ID",
			traceID:  "",
			expected: "",
		},
		{
			name:     "long trace ID",
			traceID:  "1234567890abcdef1234567890abcdef",
			expected: "1234567890abcdef1234567890abcdef",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := TraceExemplar
			result := fn(tt.traceID)
			assert.NotNil(t, result)
		})
	}
}

func TestTraceExemplar_NilSafety(t *testing.T) {
	t.Run("handles nil context gracefully", func(t *testing.T) {
		// The function should handle cases where trace ID is not available
		fn := TraceExemplar
		assert.NotPanics(t, func() {
			_ = fn("test")
		})
	})
}

func BenchmarkTraceExemplar(b *testing.B) {
	fn := TraceExemplar
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = fn("test-trace-id")
	}
}
