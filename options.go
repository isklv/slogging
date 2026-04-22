package slogging

import (
	"log"

	"github.com/Graylog2/go-gelf/gelf"
)

func NewOptions() *LoggerOptions {
	return &LoggerOptions{
		level:      defaultLevel,
		withSource: defaultWithSource,
		setDefault: defaultSetDefault,
		inGraylog:  nil,
	}
}

func (c *LoggerOptions) SetLevel(level string) *LoggerOptions {
	c.level.UnmarshalText([]byte(level))
	return c
}

func (c *LoggerOptions) WithSource(with bool) *LoggerOptions {
	c.withSource = with
	return c
}

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

func (c *LoggerOptions) SetDefault(set bool) *LoggerOptions {
	c.setDefault = set
	return c
}
