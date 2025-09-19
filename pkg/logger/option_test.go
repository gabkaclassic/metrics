package logger

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLevelOption(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantLevel slog.Level
	}{
		{"debug", "debug", slog.LevelDebug},
		{"info", "info", slog.LevelInfo},
		{"warn", "warn", slog.LevelWarn},
		{"warning", "warning", slog.LevelWarn},
		{"error", "error", slog.LevelError},
		{"unknown defaults to info", "unknown", slog.LevelInfo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &options{}
			Level(tt.input)(o)
			assert.Equal(t, tt.wantLevel, o.level)
		})
	}
}
