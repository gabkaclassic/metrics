package logger

import (
	"log/slog"
	"strings"
)

type Option func(*options)

type options struct {
	level   slog.Level
	json    bool
	console bool
	file    string
}

func JSON(enabled bool) Option {
	return func(o *options) { o.json = enabled }
}

func Console(enabled bool) Option {
	return func(o *options) { o.console = enabled }
}

func File(path string) Option {
	return func(o *options) { o.file = path }
}

func Level(level string) Option {
	return func(o *options) {
		switch strings.ToLower(level) {
		case "debug":
			o.level = slog.LevelDebug
		case "info":
			o.level = slog.LevelInfo
		case "warn", "warning":
			o.level = slog.LevelWarn
		case "error":
			o.level = slog.LevelError
		default:
			o.level = slog.LevelInfo
		}
	}
}
