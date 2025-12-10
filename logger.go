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

type LoggerOptions struct {
	level      Level
	withSource bool
	setDefault bool
	inGraylog  *gelfData
}

type gelfData struct {
	w             *gelf.Writer
	level         Level
	containerName string
}

const (
	defaultLevel      = LevelDebug
	defaultWithSource = true
	defaultSetDefault = true
)

// NewLogger opts can be
// InGraylog()
// SetLevel()
// WithSource()
// SetDefault()
func NewLogger(opts *LoggerOptions) *SLogger {
	var l *Logger

	var stdHandler Handler
	handlerOpts := &HandlerOptions{
		AddSource: opts.withSource,
		Level:     opts.level,
	}

	stdHandler = NewTextHandler(os.Stdout, handlerOpts)

	if opts.inGraylog == nil {
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

type SLogger struct {
	*slog.Logger
}

func (l *SLogger) Fatal(msg string, args ...any) {
	l.Log(context.Background(), LevelFatal, msg, args...)
	time.Sleep(1 * time.Second)
	os.Exit(1)
}
