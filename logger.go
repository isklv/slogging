package slogging

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/Graylog2/go-gelf/gelf"
	sloggraylog "github.com/samber/slog-graylog/v2"
	slogmulti "github.com/samber/slog-multi"
)

// LoggerOptions holds configuration for the logger.
type LoggerOptions struct {
	level      Level
	withSource bool
	setDefault bool
	inGraylog  *gelfData
}

// gelfData holds Graylog GELF configuration.
type gelfData struct {
	w             *gelf.Writer
	level         Level
	containerName string
}

// Default configuration values
const (
	defaultLevel      = LevelDebug
	defaultWithSource = true
	defaultSetDefault = true
)

// NewLogger creates a new SLogger with the given options.
// Usage: slogging.NewLogger(slogging.NewOptions().SetLevel("info"))
func NewLogger(opts *LoggerOptions) *SLogger {
	var l *Logger

	var stdHandler Handler
	handlerOpts := &HandlerOptions{
		AddSource: opts.withSource,
		Level:     opts.level,
	}

	stdHandler = NewTextHandler(os.Stdout, handlerOpts)

	// Safety check: ensure inGraylog and its writer are valid
	if opts.inGraylog == nil || opts.inGraylog.w == nil {
		l = New(stdHandler)
	} else {
		sloggraylog.SourceKey = "reference"
		graylogHandler := Option{
			Level:     slog.LevelDebug,
			Writer:    opts.inGraylog.w,
			Converter: sloggraylog.DefaultConverter,
			AddSource: opts.withSource,
		}.NewGraylogHandler()

		graylogHandler = graylogHandler.WithAttrs([]Attr{
			slog.String("container_name", opts.inGraylog.containerName)},
		)

		l = New(slogmulti.Fanout(stdHandler, graylogHandler))
	}

	if opts.setDefault {
		slog.SetDefault(l)
	}

	return &SLogger{
		Logger: l,
	}
}

// SLogger wraps slog.Logger with additional functionality.
type SLogger struct {
	*slog.Logger
}

// Fatal logs a message at LevelFatal and exits with code 1.
func (l *SLogger) Fatal(msg string, args ...any) {
	l.Log(context.Background(), LevelFatal, msg, args...)
	time.Sleep(1 * time.Second)
	os.Exit(1)
}
