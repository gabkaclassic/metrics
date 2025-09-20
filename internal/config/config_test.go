package config

import (
	"flag"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func resetFlags() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
}

func TestParseServerConfig(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
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
			wantAddress: "localhost:8080",
			wantLog: Log{
				Level:   "info",
				File:    "",
				Console: false,
				JSON:    true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetFlags()
			os.Args = tt.args
			cfg := ParseServerConfig()
			assert.Equal(t, tt.wantAddress, cfg.Address)
			assert.Equal(t, tt.wantLog, cfg.Log)
		})
	}
}

func TestParseAgentConfig(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
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
			wantPoll:   5 * time.Second,
			wantReport: 15 * time.Second,
			wantClient: Client{
				BaseUrl: "http://localhost:8080/update",
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
			wantPoll:   2 * time.Second,
			wantReport: 10 * time.Second,
			wantClient: Client{
				BaseUrl: "localhost:8080",
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetFlags()
			os.Args = tt.args
			cfg := ParseAgentConfig()
			assert.Equal(t, tt.wantPoll, cfg.PollInterval)
			assert.Equal(t, tt.wantReport, cfg.ReportInterval)
			assert.Equal(t, tt.wantClient, cfg.Client)
			assert.Equal(t, tt.wantLog, cfg.Log)
		})
	}
}
