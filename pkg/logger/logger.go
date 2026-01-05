package logger

import (
	"io"
	"log/slog"
	"os"
)

// LogConfig defines logger configuration parameters.
type LogConfig struct {
	// Logging level: debug, info, warn, error.
	Level string

	// Path to log file. If empty, file logging is disabled.
	File string

	// Enable logging to stdout.
	Console bool

	// Enable JSON log format. If false, text format is used.
	JSON bool
}

// SetupLogger initializes and sets the default application logger.
//
// The function configures log level, format and output destinations
// according to provided configuration and installs the logger globally
// using slog.SetDefault.
func SetupLogger(cfg LogConfig) {

	logger := new(
		JSON(cfg.JSON),
		Level(cfg.Level),
		Console(cfg.Console),
	)

	slog.SetDefault(logger)
}

// new creates a configured slog.Logger instance using functional options.
//
// Defaults:
//   - level: info
//   - format: JSON
//   - output: disabled (no writers)
//
// Intended for internal use only.
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
