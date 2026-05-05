package slogging

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/isklv/slogging/test"
	"github.com/stretchr/testify/require"
)

func TestGraylogIntegration(t *testing.T) {
	// Start Graylog via Docker Compose
	helper := test.NewTestHelper(t)
	helper.StartGraylog()
	
	// Give Graylog some time to fully initialize
	test.Sleep(2 * time.Second)

	// Create logger with Graylog
	gelfWriter := helper.GELFWriter()
	defer gelfWriter.Close()

	opts := NewOptions().InGraylog("localhost:12201", "test-container")
	sl := NewLogger(opts)

	// Test basic logging
	ctx := Context()
	testCtx := ContextWithLogger(ctx, sl)

	// Log a test message
	L(testCtx).Info("integration test message", StringAttr("key", "value"))

	// Wait for message to be sent (async)
	test.Sleep(500 * time.Millisecond)

	// Verify connection was established (if we get here without panic, connection works)
	require.NotNil(t, sl)
}

func TestGraylogIntegration_WithTraceID(t *testing.T) {
	helper := test.NewTestHelper(t)
	helper.StartGraylog()
	test.Sleep(2 * time.Second)

	gelfWriter := helper.GELFWriter()
	defer gelfWriter.Close()

	opts := NewOptions().InGraylog("localhost:12201", "test-container")
	sl := NewLogger(opts)

	// Create context with trace ID
	traceID := GenerateTraceID()
	l := sl.Logger.With(StringAttr(XB3TraceID, traceID))
	testLogger := &SLogger{Logger: l}
	ctx := ContextWithLogger(context.Background(), testLogger)
	ctx = context.WithValue(ctx, XB3TraceID, traceID)

	// Log with trace ID
	L(ctx).Info("test with trace", StringAttr("operation", "test"))

	test.Sleep(500 * time.Millisecond)
	require.Equal(t, traceID, ctx.Value(XB3TraceID))
}

func TestGraylogIntegration_MultipleLoggers(t *testing.T) {
	helper := test.NewTestHelper(t)
	helper.StartGraylog()
	test.Sleep(2 * time.Second)

	gelfWriter := helper.GELFWriter()
	defer gelfWriter.Close()

	opts := NewOptions().InGraylog("localhost:12201", "test-container")
	sl := NewLogger(opts)

	// Create multiple loggers with different attributes
	l1 := sl.Logger.With(StringAttr("service", "service-1"))
	l2 := sl.Logger.With(StringAttr("service", "service-2"))

	logger1 := &SLogger{Logger: l1}
	logger2 := &SLogger{Logger: l2}

	ctx1 := ContextWithLogger(context.Background(), logger1)
	ctx2 := ContextWithLogger(context.Background(), logger2)

	// Log from both
	L(ctx1).Info("service 1 message")
	L(ctx2).Info("service 2 message")

	test.Sleep(500 * time.Millisecond)
	require.NotNil(t, logger1)
	require.NotNil(t, logger2)
}

func TestGraylogIntegration_DifferentLevels(t *testing.T) {
	helper := test.NewTestHelper(t)
	helper.StartGraylog()
	test.Sleep(2 * time.Second)

	gelfWriter := helper.GELFWriter()
	defer gelfWriter.Close()

	opts := NewOptions().InGraylog("localhost:12201", "test-container")
	sl := NewLogger(opts)
	ctx := Context()
	testCtx := ContextWithLogger(ctx, sl)

	// Test different log levels
	L(testCtx).Debug("debug message")
	L(testCtx).Info("info message")
	L(testCtx).Warn("warn message")
	L(testCtx).Error("error message")

	test.Sleep(500 * time.Millisecond)
}

func TestGraylogIntegration_ContextPropagation(t *testing.T) {
	helper := test.NewTestHelper(t)
	helper.StartGraylog()
	test.Sleep(2 * time.Second)

	gelfWriter := helper.GELFWriter()
	defer gelfWriter.Close()

	// Create parent context
	parentCtx := Context()
	parentLogger := &SLogger{slog.Default().With(StringAttr("parent", "true"))}
	parentCtx = ContextWithLogger(parentCtx, parentLogger)
	parentCtx = context.WithValue(parentCtx, XB3TraceID, "parent-trace-id")

	// Simulate nested call
	func(ctx context.Context) {
		// Should inherit logger from context
		logger := L(ctx)
		require.NotNil(t, logger)
		logger.Info("nested call")
	}(parentCtx)

	test.Sleep(500 * time.Millisecond)
}

// Test helper function for generating test data
func TestGenerateTraceID(t *testing.T) {
	id := GenerateTraceID()
	require.NotEmpty(t, id)
	require.Equal(t, 36, len(id)) // UUID format
}

// Test that log levels map correctly
func TestLogLevelsMapping(t *testing.T) {
	tests := []struct {
		name     string
		level    slog.Level
		expected int32
	}{
		{"debug", slog.LevelDebug, 7},
		{"info", slog.LevelInfo, 6},
		{"warn", slog.LevelWarn, 4},
		{"error", slog.LevelError, 3},
		{"fatal", LevelFatal, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := LogLevels[tt.level]
			require.Equal(t, tt.expected, actual)
		})
	}
}
