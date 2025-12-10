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

const (
	LevelDebug = slog.LevelDebug
	LevelInfo  = slog.LevelInfo
	LevelWarn  = slog.LevelWarn
	LevelError = slog.LevelError
	LevelFatal = slog.Level(12)
)

type (
	Logger         = slog.Logger
	Level          = slog.Level
	Record         = slog.Record
	Handler        = slog.Handler
	Attr           = slog.Attr
	HandlerOptions = slog.HandlerOptions
)

var (
	New            = slog.New
	NewTextHandler = slog.NewTextHandler
)

func IntAttr[T constraints.Integer](key string, value T) Attr {
	return slog.Int(key, int(value))
}

func FloatAttr[T constraints.Float](key string, value T) Attr {
	return slog.Float64(key, float64(value))
}

func TimeAttr(key string, time time.Time) Attr {
	return slog.String(key, time.Format("2006-01-02 15:04:05"))
}

func ErrAttr(err error) Attr {
	return slog.String("error", err.Error())
}

func StringAttr(key string, value string) Attr {
	return slog.String(key, value)
}

func AnyAttr(key string, s interface{}) Attr {
	v := reflect.ValueOf(s)

	if v.Kind() == reflect.Ptr && !v.IsNil() {
		s = v.Elem().Interface()
	}

	return slog.Any(key, s)
}

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
		slog.String("url", r.Request.URL.String()),
		slog.String("method", r.Request.Method),
		slog.Int("statusCode", r.StatusCode),
		slog.String("headers", string(headers)),
		slog.String("body", string(body)),
		slog.Int64("duration", duration.Milliseconds()),
	})
}

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
		slog.String("method", r.Method),
		slog.String("url", r.URL.String()),
		slog.String("headers", string(headers)),
		slog.String("body", string(body)),
	})
}

func checkHeaderAuth(header string) string {
	parts := strings.SplitN(header, " ", 2)
	if len(parts) == 1 {
		return maskToken(parts[0])
	}

	scheme, data := parts[0], parts[1]
	return scheme + " " + maskToken(data)
}

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
