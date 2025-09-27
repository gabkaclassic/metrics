package logger

import (
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name      string
		console   bool
		json      bool
		file      string
		wantPanic bool
	}{
		{"console json", true, true, "", false},
		{"console text", true, false, "", false},
		{"file output", false, true, "test.log", false},
		{"console and file", true, false, "test.log", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				assert.Panics(t, func() { new(JSON(tt.json), Console(tt.console), File(tt.file)) })
				return
			}
			logger := new(JSON(tt.json), Console(tt.console), File(tt.file))
			assert.NotNil(t, logger)
		})
	}

	_ = os.Remove("test.log")
}

func TestSetupLogger(t *testing.T) {
	tests := []struct {
		name        string
		cfg         LogConfig
		wantJSON    bool
		wantConsole bool
		wantLevel   slog.Level
	}{
		{
			name: "console info json",
			cfg: LogConfig{
				Level:   "info",
				Console: true,
				JSON:    true,
			},
			wantJSON:    true,
			wantConsole: true,
			wantLevel:   slog.LevelInfo,
		},
		{
			name: "console warn text",
			cfg: LogConfig{
				Level:   "warn",
				Console: true,
				JSON:    false,
			},
			wantJSON:    false,
			wantConsole: true,
			wantLevel:   slog.LevelWarn,
		},
		{
			name: "no console defaults",
			cfg: LogConfig{
				Level:   "debug",
				Console: false,
				JSON:    true,
			},
			wantJSON:    true,
			wantConsole: false,
			wantLevel:   slog.LevelDebug,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetupLogger(tt.cfg)
			logger := slog.Default()
			assert.NotNil(t, logger)
			h := logger.Handler()
			switch tt.wantJSON {
			case true:
				_, ok := h.(*slog.JSONHandler)
				assert.True(t, ok)
			case false:
				_, ok := h.(*slog.TextHandler)
				assert.True(t, ok)
			}
		})
	}
}
