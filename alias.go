package slogging

import (
	"bytes"
	"encoding/json"
	"golang.org/x/exp/constraints"
	"io"
	"log/slog"
	"net/http"
	"reflect"
	"strings"
	"time"
)

// Log levels
const (
	LevelDebug = slog.LevelDebug
	LevelInfo  = slog.LevelInfo
	LevelWarn  = slog.LevelWarn
	LevelError = slog.LevelError
	LevelFatal = slog.Level(12)
)

// Type aliases for slog
type (
	Logger         = slog.Logger
	Level          = slog.Level
	Record         = slog.Record
	Handler        = slog.Handler
	Attr           = slog.Attr
	HandlerOptions = slog.HandlerOptions
)

// Function aliases for slog
var (
	New            = slog.New
	NewTextHandler = slog.NewTextHandler
)

// IntAttr creates an integer attribute with type constraint.
func IntAttr[T constraints.Integer](key string, value T) Attr {
	return slog.Int(key, int(value))
}

// FloatAttr creates a float attribute with type constraint.
func FloatAttr[T constraints.Float](key string, value T) Attr {
	return slog.Float64(key, float64(value))
}

// TimeAttr creates a time attribute formatted as "2006-01-02 15:04:05".
func TimeAttr(key string, time time.Time) Attr {
	return slog.String(key, time.Format("2006-01-02 15:04:05"))
}

// ErrAttr creates an error attribute with key "error".
func ErrAttr(err error) Attr {
	return slog.String("error", err.Error())
}

// StringAttr creates a string attribute.
func StringAttr(key string, value string) Attr {
	return slog.String(key, value)
}

// AnyAttr creates an attribute from any value using reflection.
func AnyAttr(key string, s interface{}) Attr {
	v := reflect.ValueOf(s)

	if v.Kind() == reflect.Ptr && !v.IsNil() {
		s = v.Elem().Interface()
	}

	return slog.Any(key, s)
}

// ResponseAttr extracts attributes from HTTP response including status, headers, body and duration.
func ResponseAttr(r *http.Response, start time.Time) []any {
	if r == nil {
		slog.Error("response is nil")
		return []any{}
	}

	var body []byte
	if r.Body != nil {
		body, _ = io.ReadAll(r.Body)
		r.Body = io.NopCloser(bytes.NewBuffer(body))
	}

	headers, _ := json.Marshal(r.Header)

	duration := time.Since(start)
	return getReqAttrsAsAny([]Attr{
		StringAttr("url", r.Request.URL.String()),
		StringAttr("method", r.Request.Method),
		IntAttr("statusCode", r.StatusCode),
		StringAttr("headers", string(headers)),
		StringAttr("body", string(body)),
		IntAttr("duration", duration.Milliseconds()),
	})
}

// RequestAttr extracts attributes from HTTP request including method, url, headers and body.
// Masks Authorization headers for security.
func RequestAttr(r *http.Request) []any {
	if r == nil {
		slog.Error("request is nil")
		return []any{}
	}

	var body []byte
	if r.Body != nil {
		body, _ = io.ReadAll(r.Body)
		r.Body = io.NopCloser(bytes.NewBuffer(body))
	}

	logHeaders := make(http.Header)
	for k, v := range r.Header {
		logHeaders[k] = v
	}

	if authHeader := logHeaders.Get("Authorization"); authHeader != "" {
		logHeaders.Set("Authorization", checkHeaderAuth(authHeader))
	}

	headers, _ := json.Marshal(logHeaders)

	return getReqAttrsAsAny([]Attr{
		StringAttr("method", r.Method),
		StringAttr("url", r.URL.String()),
		StringAttr("headers", string(headers)),
		StringAttr("body", string(body)),
	})
}

// checkHeaderAuth parses and masks authentication header.
func checkHeaderAuth(header string) string {
	parts := strings.SplitN(header, " ", 2)
	if len(parts) == 1 {
		return maskToken(parts[0])
	}

	scheme, data := parts[0], parts[1]
	return scheme + " " + maskToken(data)
}

// maskToken masks token keeping first 2 and last 2 characters.
func maskToken(token string) string {
	const mask = "***"
	if len(token) <= 4 {
		return mask
	}
	return token[:2] + mask + token[len(token)-2:]
}

func getReqAttrsAsAny(reqAttrs []Attr) []any {
	args := make([]any, len(reqAttrs))
	for i, attr := range reqAttrs {
		args[i] = attr
	}
	return args
}
