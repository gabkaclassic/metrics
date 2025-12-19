package logger

import (
	"log/slog"
	"strings"
)

// Option represents a functional option for logger configuration.
type Option func(*options)

// options holds resolved logger configuration values.
type options struct {
	level   slog.Level
	json    bool
	console bool
	file    string
}

// JSON enables or disables JSON log format.
func JSON(enabled bool) Option {
	return func(o *options) { o.json = enabled }
}

// Console enables or disables logging to stdout.
func Console(enabled bool) Option {
	return func(o *options) { o.console = enabled }
}

// File enables logging to the specified file path.
//
// The file is opened in append mode and created if it does not exist.
func File(path string) Option {
	return func(o *options) { o.file = path }
}

// Level sets the logging level.
//
// Supported values:
//   - debug
//   - info
//   - warn, warning
//   - error
//
// Unknown values default to info.
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
