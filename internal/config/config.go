package config

import (
	"flag"
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

func ParseServerConfig() *Server {
	var cfg Server

	address := flag.String("address", "0.0.0.0:8080", "HTTP server address")

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

	pollInterval := flag.Duration("poll-interval", 2*time.Second, "Metrics polling interval")
	reportInterval := flag.Duration("report-interval", 10*time.Second, "Metrics reporting interval")
	serverAddress := flag.String("report-url", "http://0.0.0.0:8080/update", "Server HTTP base URL")
	retries := flag.Int("report-retries", 3, "Max update metrics retries")
	timeout := flag.Duration("report-timeout", 3*time.Second, "Metrics update timeout")

	// Logging
	logLevel := flag.String("log-level", "info", "Logging level")
	logFile := flag.String("log-file", "", "Log file path")
	logConsole := flag.Bool("log-console", false, "Enable console logging")
	logJSON := flag.Bool("log-json", true, "Enable JSON output for logs")

	flag.Parse()

	cfg.PollInterval = *pollInterval
	cfg.ReportInterval = *reportInterval
	cfg.Client.BaseUrl = *serverAddress
	cfg.Client.Retries = *retries
	cfg.Client.Timeout = *timeout
	cfg.Log = Log{
		Level:   *logLevel,
		File:    *logFile,
		Console: *logConsole,
		JSON:    *logJSON,
	}

	return &cfg
}
