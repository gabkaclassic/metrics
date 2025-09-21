package config

import (
	"flag"
	"net/url"
	"strings"
	"time"
)

type (
	Server struct {
		Address string
		Log     Log
	}
	Agent struct {
		PollInterval   time.Duration
		ReportInterval time.Duration
		Client         Client
		Log            Log
	}
	Client struct {
		BaseUrl string
		Timeout time.Duration
		Retries int
	}
	Log struct {
		Level   string
		File    string
		Console bool
		JSON    bool
	}
)

func ensureURL(addr string) string {
	if !strings.Contains(addr, "://") {
		addr = "http://" + addr
	}
	u, err := url.Parse(addr)
	if err != nil {
		return addr
	}
	return u.String()
}

func ParseServerConfig() *Server {
	var cfg Server

	address := flag.String("a", "localhost:8080", "HTTP server address")

	// Logging
	logLevel := flag.String("log-level", "info", "Logging level")
	logFile := flag.String("log-file", "", "Log file path")
	logConsole := flag.Bool("log-console", false, "Enable console logging")
	logJSON := flag.Bool("log-json", true, "Enable JSON output for logs")

	flag.Parse()

	cfg.Address = *address
	cfg.Log = Log{
		Level:   *logLevel,
		File:    *logFile,
		Console: *logConsole,
		JSON:    *logJSON,
	}

	return &cfg
}

func ParseAgentConfig() *Agent {
	var cfg Agent

	pollInterval := flag.Uint("p", 2, "Metrics polling interval (seconds)")
	reportInterval := flag.Uint("r", 10, "Metrics reporting interval (seconds)")
	serverAddress := flag.String("a", "http://localhost:8080", "Server HTTP base URL")
	retries := flag.Int("report-retries", 3, "Max update metrics retries")
	timeout := flag.Uint("report-timeout", 3, "Metrics update timeout (seconds)")

	// Logging
	logLevel := flag.String("log-level", "info", "Logging level")
	logFile := flag.String("log-file", "", "Log file path")
	logConsole := flag.Bool("log-console", false, "Enable console logging")
	logJSON := flag.Bool("log-json", true, "Enable JSON output for logs")

	flag.Parse()

	cfg.PollInterval = time.Duration(*pollInterval) * time.Second
	cfg.ReportInterval = time.Duration(*reportInterval) * time.Second
	cfg.Client.BaseUrl = ensureURL(*serverAddress)
	cfg.Client.Retries = *retries
	cfg.Client.Timeout = time.Duration(*timeout) * time.Second
	cfg.Log = Log{
		Level:   *logLevel,
		File:    *logFile,
		Console: *logConsole,
		JSON:    *logJSON,
	}

	return &cfg
}
