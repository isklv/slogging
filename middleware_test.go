package slogging

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// Concurrent safety tests

func TestConcurrentLogging(t *testing.T) {
	sl := NewLogger(NewOptions())
	ctx := Context()
	testCtx := ContextWithLogger(ctx, sl)

	var wg sync.WaitGroup
	wg.Add(100)

	for i := 0; i < 100; i++ {
		go func(id int) {
			defer wg.Done()
			L(testCtx).Info("concurrent log", IntAttr("id", id))
		}(i)
	}

	wg.Wait()
}

// Test RequestAttr and ResponseAttr helpers
func TestRequestAttr_MasksAuthorization(t *testing.T) {
	body := io.NopCloser(strings.NewReader("test body"))
	req := httptest.NewRequest("POST", "/test", body)
	req.Header.Set("Authorization", "Bearer secret-token-12345")
	req.Header.Set("Content-Type", "application/json")

	attrs := RequestAttr(req)
	
	// Convert attrs to string for checking
	attrStr := fmt.Sprintf("%v", attrs)
	
	// Should not contain full token
	require.NotContains(t, attrStr, "secret-token-12345")
	// Should contain masked version (se**45)
	require.Contains(t, attrStr, "se")
	require.Contains(t, attrStr, "45")
	require.Contains(t, attrStr, "***")
}

func TestRequestAttr_NilRequest(t *testing.T) {
	attrs := RequestAttr(nil)
	require.Empty(t, attrs)
}

func TestResponseAttr(t *testing.T) {
	start := time.Now()
	time.Sleep(10 * time.Millisecond)

	resp := &http.Response{
		StatusCode: 200,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader("response body")),
		Request:    &http.Request{URL: &url.URL{Path: "/test"}},
	}
	resp.Header.Set("Content-Type", "application/json")

	attrs := ResponseAttr(resp, start)
	require.NotEmpty(t, attrs)
}

func TestResponseAttr_NilResponse(t *testing.T) {
	attrs := ResponseAttr(nil, time.Now())
	require.Empty(t, attrs)
}

// Test token masking helper
func TestMaskToken(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"long token", "secret-token-12345", "se***45"},
		{"short token", "ab", "***"},
		{"exact 4 chars", "abcd", "***"},
		{"5 chars", "abcde", "ab***de"},
		{"Bearer prefix", "Bearer mytoken123", "Bearer my***23"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := checkHeaderAuth(tt.input)
			require.Equal(t, tt.expected, result)
		})
	}
}

// Test context operations
func TestContextLoggerPropagation(t *testing.T) {
	sl := NewLogger(NewOptions())
	ctx := Context()
	
	// Set logger in context
	ctx = ContextWithLogger(ctx, sl)
	
	// Retrieve it
	retrieved := L(ctx)
	require.NotNil(t, retrieved)
	
	// Create child context
	childCtx := context.WithValue(ctx, "key", "value")
	childLogger := L(childCtx)
	require.NotNil(t, childLogger)
}

func TestContextWithoutLogger(t *testing.T) {
	ctx := context.Background()
	logger := L(ctx)
	require.NotNil(t, logger)
}

func TestContextWithTraceID(t *testing.T) {
	ctx := Context()
	traceID, ok := ctx.Value(XB3TraceID).(string)
	require.True(t, ok)
	require.NotEmpty(t, traceID)
	require.Equal(t, 36, len(traceID))
}
