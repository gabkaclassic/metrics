package config

import (
	"flag"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func resetFlags() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
}

func TestEnsureURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "already has http scheme",
			input:    "http://localhost:8080",
			expected: "http://localhost:8080",
		},
		{
			name:     "already has https scheme",
			input:    "https://example.com:443",
			expected: "https://example.com:443",
		},
		{
			name:     "without scheme, add http",
			input:    "localhost:8080",
			expected: "http://localhost:8080",
		},
		{
			name:     "hostname without port",
			input:    "example.com",
			expected: "http://example.com",
		},
		{
			name:     "invalid url",
			input:    "http://[::1:8080",
			expected: "http://[::1:8080",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "http:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ensureURL(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseServerConfig(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		env         map[string]string
		wantAddress string
		wantLog     Log
	}{
		{
			name: "custom flags",
			args: []string{
				"cmd",
				"-a=127.0.0.1:9000",
				"-log-level=debug",
				"-log-file=test.log",
				"-log-console=true",
				"-log-json=false",
			},
			env:         map[string]string{},
			wantAddress: "127.0.0.1:9000",
			wantLog: Log{
				Level:   "debug",
				File:    "test.log",
				Console: true,
				JSON:    false,
			},
		},
		{
			name:        "default values",
			args:        []string{"cmd"},
			env:         map[string]string{},
			wantAddress: "localhost:8080",
			wantLog: Log{
				Level:   "info",
				File:    "",
				Console: false,
				JSON:    true,
			},
		},
		{
			name: "values from env",
			args: []string{"cmd"},
			env: map[string]string{
				"ADDRESS":     "envhost:5555",
				"LOG_LEVEL":   "warn",
				"LOG_FILE":    "env.log",
				"LOG_CONSOLE": "true",
				"LOG_JSON":    "false",
			},
			wantAddress: "envhost:5555",
			wantLog: Log{
				Level:   "warn",
				File:    "env.log",
				Console: true,
				JSON:    false,
			},
		},
		{
			name: "env overridden by flags",
			args: []string{
				"cmd",
				"-a=flaghost:9999",
				"-log-level=error",
			},
			env: map[string]string{
				"ADDRESS":   "envhost:5555",
				"LOG_LEVEL": "warn",
			},
			wantAddress: "flaghost:9999",
			wantLog: Log{
				Level:   "error",
				File:    "",
				Console: false,
				JSON:    true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetFlags()
			resetEnv("ADDRESS", "LOG_LEVEL", "LOG_FILE", "LOG_CONSOLE", "LOG_JSON")

			for k, v := range tt.env {
				_ = os.Setenv(k, v)
			}

			os.Args = tt.args
			cfg, err := ParseServerConfig()
			require.NoError(t, err)

			assert.Equal(t, tt.wantAddress, cfg.Address)
			assert.Equal(t, tt.wantLog, cfg.Log)
		})
	}
}

func TestParseAgentConfig(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		env        map[string]string
		wantPoll   time.Duration
		wantReport time.Duration
		wantClient Client
		wantLog    Log
	}{
		{
			name: "custom flags",
			args: []string{
				"cmd",
				"-p=5",
				"-r=15",
				"-a=http://localhost:8080/update",
				"-report-retries=5",
				"-report-timeout=4",
				"-log-level=warn",
				"-log-file=test_agent.log",
				"-log-console=true",
				"-log-json=true",
			},
			env:        map[string]string{},
			wantPoll:   5 * time.Second,
			wantReport: 15 * time.Second,
			wantClient: Client{
				BaseURL: "http://localhost:8080/update",
				Retries: 5,
				Timeout: 4 * time.Second,
			},
			wantLog: Log{
				Level:   "warn",
				File:    "test_agent.log",
				Console: true,
				JSON:    true,
			},
		},
		{
			name:       "default values",
			args:       []string{"cmd"},
			env:        map[string]string{},
			wantPoll:   2 * time.Second,
			wantReport: 10 * time.Second,
			wantClient: Client{
				BaseURL: "http://localhost:8080",
				Retries: 3,
				Timeout: 3 * time.Second,
			},
			wantLog: Log{
				Level:   "info",
				File:    "",
				Console: false,
				JSON:    true,
			},
		},
		{
			name: "values from env",
			args: []string{"cmd"},
			env: map[string]string{
				"POLL_INTERVAL":   "7",
				"REPORT_INTERVAL": "20",
				"ADDRESS":         "envhost:7777",
				"RETRIES":         "8",
				"TIMEOUT":         "9",
				"LOG_LEVEL":       "debug",
				"LOG_FILE":        "env_agent.log",
				"LOG_CONSOLE":     "true",
				"LOG_JSON":        "false",
			},
			wantPoll:   7 * time.Second,
			wantReport: 20 * time.Second,
			wantClient: Client{
				BaseURL: "http://envhost:7777",
				Retries: 8,
				Timeout: 9 * time.Second,
			},
			wantLog: Log{
				Level:   "debug",
				File:    "env_agent.log",
				Console: true,
				JSON:    false,
			},
		},
		{
			name: "env overridden by flags",
			args: []string{
				"cmd",
				"-p=3",
				"-a=http://flaghost:9999",
				"-log-level=error",
			},
			env: map[string]string{
				"POLL_INTERVAL": "12",
				"ADDRESS":       "envhost:8888",
				"LOG_LEVEL":     "warn",
			},
			wantPoll:   3 * time.Second,
			wantReport: 10 * time.Second,
			wantClient: Client{
				BaseURL: "http://flaghost:9999",
				Retries: 3,
				Timeout: 3 * time.Second,
			},
			wantLog: Log{
				Level:   "error",
				File:    "",
				Console: false,
				JSON:    true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetFlags()
			resetEnv("POLL_INTERVAL", "REPORT_INTERVAL", "ADDRESS", "RETRIES", "TIMEOUT", "LOG_LEVEL", "LOG_FILE", "LOG_CONSOLE", "LOG_JSON")

			// установить ENV
			for k, v := range tt.env {
				_ = os.Setenv(k, v)
			}

			os.Args = tt.args
			cfg, err := ParseAgentConfig()
			require.NoError(t, err)

			assert.Equal(t, tt.wantPoll, cfg.PollInterval)
			assert.Equal(t, tt.wantReport, cfg.ReportInterval)
			assert.Equal(t, tt.wantClient, cfg.Client)
			assert.Equal(t, tt.wantLog, cfg.Log)
		})
	}
}

func resetEnv(vars ...string) {
	for _, v := range vars {
		_ = os.Unsetenv(v)
	}
}
