package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gabkaclassic/metrics/internal/agent"
	"github.com/gabkaclassic/metrics/internal/config"
	"github.com/gabkaclassic/metrics/pkg/httpclient"
	"github.com/gabkaclassic/metrics/pkg/logger"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := config.ParseAgentConfig()
	if err != nil {
		return fmt.Errorf("failed to parse agent configuration: %w", err)
	}

	logger.SetupLogger(logger.LogConfig(cfg.Log))

	client := httpclient.NewClient(
		httpclient.BaseURL(cfg.Client.BaseURL),
		httpclient.Timeout(cfg.Client.Timeout),
		httpclient.MaxRetries(cfg.Client.Retries),
	)

	agent, err := agent.NewAgent(
		client, cfg.BatchesEnabled, cfg.SignKey, cfg.RateLimit,
	)
	if err != nil {
		return fmt.Errorf("failed to initialize agent: %w", err)
	}

	startAgent(cfg.PollInterval, cfg.ReportInterval, agent)

	return nil
}

func startAgent(pollInterval time.Duration, reportInterval time.Duration, agent *agent.MetricsAgent) {
	pollTicker := time.NewTicker(pollInterval)
	reportTicker := time.NewTicker(reportInterval)
	defer pollTicker.Stop()
	defer reportTicker.Stop()

	done := make(chan struct{})

	go func() {
		for {
			select {
			case <-pollTicker.C:
				agent.Poll()
				slog.Info("Poll completed")
			case <-done:
				return
			}
		}
	}()

	go func() {
		for {
			select {
			case <-reportTicker.C:
				if err := agent.Report(); err != nil {
					slog.Error(
						"Report error",
						slog.String("error", err.Error()),
					)
				} else {
					slog.Info("Report completed")
				}
			case <-done:
				return
			}
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	slog.Debug("Shutting down...")
	close(done)
	time.Sleep(100 * time.Millisecond)
}
