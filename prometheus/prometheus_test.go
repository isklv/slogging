package prometheus

import (
	"context"
	"testing"

	"github.com/isklv/slogging"
	"github.com/stretchr/testify/assert"
)

func TestTraceExemplar(t *testing.T) {
	t.Run("returns trace exemplar function", func(t *testing.T) {
		fn := TraceExemplar
		assert.NotNil(t, fn)
	})

	t.Run("function signature", func(t *testing.T) {
		// TraceExemplar should return a function that takes context and returns exemplar
		fn := TraceExemplar
		ctx := context.Background()
		result := fn(ctx)
		// Returns nil when no trace ID in context
		assert.Nil(t, result)
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
			ctx := context.WithValue(context.Background(), slogging.XB3TraceID, tt.traceID)
			fn := TraceExemplar
			result := fn(ctx)
			if tt.traceID == "" {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				if result != nil {
					assert.Equal(t, tt.expected, result[slogging.XB3TraceID])
				}
			}
		})
	}
}

func TestTraceExemplar_NilSafety(t *testing.T) {
	t.Run("handles nil context gracefully", func(t *testing.T) {
		// The function should handle cases where trace ID is not available
		fn := TraceExemplar
		assert.NotPanics(t, func() {
			_ = fn(context.Background())
		})
	})
}

func BenchmarkTraceExemplar(b *testing.B) {
	fn := TraceExemplar
	ctx := context.WithValue(context.Background(), slogging.XB3TraceID, "test-trace-id")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = fn(ctx)
	}
}
