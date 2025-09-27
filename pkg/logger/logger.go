package logger

import (
	"io"
	"log/slog"
	"os"
)

type LogConfig struct {
	Level   string
	File    string
	Console bool
	JSON    bool
}

func SetupLogger(cfg LogConfig) {

	logger := new(
		JSON(cfg.JSON),
		Level(cfg.Level),
		Console(cfg.Console),
	)

	slog.SetDefault(logger)
}

func new(opts ...Option) *slog.Logger {
	o := &options{
		level:   slog.LevelInfo,
		json:    true,
		console: false,
		file:    "",
	}

	for _, opt := range opts {
		opt(o)
	}

	var writers []io.Writer
	if o.console {
		writers = append(writers, os.Stdout)
	}
	if o.file != "" {
		f, err := os.OpenFile(o.file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err == nil {
			writers = append(writers, f)
		} else {
			slog.Warn("Cannot open log file, logging to stdout only", "error", err)
		}
	}

	writer := io.MultiWriter(writers...)

	var handler slog.Handler
	if o.json {
		handler = slog.NewJSONHandler(writer, &slog.HandlerOptions{Level: o.level})
	} else {
		handler = slog.NewTextHandler(writer, &slog.HandlerOptions{Level: o.level})
	}

	return slog.New(handler)
}
