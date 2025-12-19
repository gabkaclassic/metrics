// Package logger provides application-wide logging configuration.
//
// The package is a thin wrapper around log/slog and allows configuring:
//   - Log level
//   - Output format (JSON or text)
//   - Output destinations (stdout and/or file)
//
// Logger configuration is applied globally via slog.SetDefault
// and should be performed once during application startup.
package logger
