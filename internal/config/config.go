package config

import (
	"flag"
	"time"
)

type (
	Server struct {
		Address string
	}
	Agent struct {
		PollInterval   time.Duration
		ReportInterval time.Duration
		Client         Client
	}
	Client struct {
		BaseUrl string
		Timeout time.Duration
		Retries int
	}
)

func ParseServerConfig() *Server {
	var cfg Server

	address := flag.String("address", "0.0.0.0:8080", "HTTP server address")

	flag.Parse()

	cfg.Address = *address

	return &cfg
}

func ParseAgentConfig() *Agent {
	var cfg Agent

	pollInterval := flag.Duration("poll-interval", 2*time.Second, "Metrics polling interval")
	reportInterval := flag.Duration("report-interval", 10*time.Second, "Metrics polling interval")
	serverAddress := flag.String("report-url", "http://0.0.0.0:8080", "Server HTTP base URL")
	retries := flag.Int("report-retries", 3, "Max update metrics retries")
	timeout := flag.Duration("report-timeout", 3*time.Second, "Metrics updating interval")

	flag.Parse()

	cfg.PollInterval = *pollInterval
	cfg.ReportInterval = *reportInterval
	cfg.Client.BaseUrl = *serverAddress
	cfg.Client.Retries = *retries
	cfg.Client.Timeout = *timeout

	return &cfg
}
