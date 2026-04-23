package slogging

import (
	"log"

	"github.com/Graylog2/go-gelf/gelf"
)

// NewOptions creates new LoggerOptions with default values.
func NewOptions() *LoggerOptions {
	return &LoggerOptions{
		level:      defaultLevel,
		withSource: defaultWithSource,
		setDefault: defaultSetDefault,
		inGraylog:  nil,
	}
}

// SetLevel sets the log level (debug, info, warn, error, fatal).
func (c *LoggerOptions) SetLevel(level string) *LoggerOptions {
	c.level.UnmarshalText([]byte(level))
	return c
}

// WithSource enables or disables source code location in logs.
func (c *LoggerOptions) WithSource(with bool) *LoggerOptions {
	c.withSource = with
	return c
}

// InGraylog configures UDP Graylog GELF output.
// Empty URL is ignored. Connection errors are logged but don't crash the app.
// Returns the options for chaining.
func (c *LoggerOptions) InGraylog(graylogURL, containerName string) *LoggerOptions {
	if graylogURL == "" {
		// Если URL пустой, просто не включаем graylog
		return c
	}

	w, err := gelf.NewWriter(graylogURL)
	if err != nil {
		// Логгируем ошибку в stderr, но не паникуем
		log.Printf("slogging: failed to connect to graylog %s: %v", graylogURL, err)
		return c
	}

	c.inGraylog = &gelfData{
		w:             w,
		containerName: containerName,
	}

	return c
}

// SetDefault sets whether this logger should be set as the default slog logger.
func (c *LoggerOptions) SetDefault(set bool) *LoggerOptions {
	c.setDefault = set
	return c
}
