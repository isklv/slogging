package slogging

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"github.com/Graylog2/go-gelf/gelf"
	slogcommon "github.com/samber/slog-common"
	sloggraylog "github.com/samber/slog-graylog/v2"
)

// Converter converts slog record to Graylog extra fields.
type Converter func(addSource bool, replaceAttr func(groups []string, a slog.Attr) slog.Attr, loggerAttr []slog.Attr, groups []string, record *slog.Record) (extra map[string]any)

// Option configures Graylog handler.
type Option struct {
	// Level sets the minimum log level (default: debug).
	Level slog.Leveler

	// Writer is the Graylog GELF writer connection.
	Writer *gelf.Writer

	// Converter customizes JSON payload builder.
	Converter Converter
	// AttrFromContext extracts additional attributes from context.
	AttrFromContext []func(ctx context.Context) []slog.Attr

	// AddSource includes source code location in logs (see slog.HandlerOptions).
	AddSource bool
	// ReplaceAttr transforms attributes before logging.
	ReplaceAttr func(groups []string, a slog.Attr) slog.Attr

	// hostname is internal, auto-detected.
	hostname string
}

// NewGraylogHandler creates a new Graylog GELF handler.
func (o Option) NewGraylogHandler() slog.Handler {
	if o.Level == nil {
		o.Level = LevelDebug
	}

	if o.AttrFromContext == nil {
		o.AttrFromContext = []func(ctx context.Context) []slog.Attr{}
	}

	if o.Converter == nil {
		o.Converter = sloggraylog.DefaultConverter
	}

	if hostname, err := os.Hostname(); err == nil {
		o.hostname = hostname
	}

	return &GraylogHandler{
		option: o,
		attrs:  []slog.Attr{},
		groups: []string{},
	}
}

var _ slog.Handler = (*GraylogHandler)(nil)

// GraylogHandler implements slog.Handler for Graylog GELF output.
type GraylogHandler struct {
	option Option
	attrs  []slog.Attr
	groups []string
}

func (h *GraylogHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.option.Level.Level()
	//return level >= h.option.Level.Level() || level == LevelFatal
}

func (h *GraylogHandler) Handle(ctx context.Context, record slog.Record) error {
	fromContext := slogcommon.ContextExtractor(ctx, h.option.AttrFromContext)
	extra := h.option.Converter(h.option.AddSource, h.option.ReplaceAttr, append(h.attrs, fromContext...), h.groups, &record)

	msg := &gelf.Message{
		Version:  "1.1",
		Host:     h.option.hostname,
		Short:    short(&record),
		TimeUnix: float64(record.Time.Unix()),
		Level:    LogLevels[record.Level],
		Extra:    extra,
	}

	// non-blocking with nil check and error ignoring
	if h.option.Writer != nil {
		go func() {
			// Ignore errors - UDP may fail if Graylog is down,
			// but we don't want to crash the app
			_ = h.option.Writer.WriteMessage(msg)
		}()
	}

	return nil
}

func (h *GraylogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &GraylogHandler{
		option: h.option,
		attrs:  slogcommon.AppendAttrsToGroup(h.groups, h.attrs, attrs...),
		groups: h.groups,
	}
}

func (h *GraylogHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}

	return &GraylogHandler{
		option: h.option,
		attrs:  h.attrs,
		groups: append(h.groups, name),
	}
}

func short(record *slog.Record) string {
	msg := strings.TrimSpace(record.Message)
	if i := strings.IndexRune(msg, '\n'); i > 0 {
		return msg[:i]
	}

	return msg
}

// XB3TraceID is the context key for B3 trace ID.
const (
	XB3TraceID = "X-B3-TraceId"
)

// LogLevels maps slog levels to Graylog GELF levels.
var LogLevels = map[slog.Level]int32{
	slog.LevelDebug: 7,
	slog.LevelInfo:  6,
	slog.LevelWarn:  4,
	slog.LevelError: 3,
	LevelFatal:      2,
}
