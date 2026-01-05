package main

import (
	"context"
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
		client, cfg.BatchesEnabled, cfg.SignKey, cfg.RateLimit, cfg.BatchSize,
	)
	if err != nil {
		return fmt.Errorf("failed to initialize agent: %w", err)
	}

	startAgent(cfg.PollInterval, cfg.ReportInterval, agent)

	return nil
}

func startAgent(pollInterval, reportInterval time.Duration, agent *agent.MetricsAgent) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pollTicker := time.NewTicker(pollInterval)
	reportTicker := time.NewTicker(reportInterval)
	defer pollTicker.Stop()
	defer reportTicker.Stop()

	for {
		select {
		case <-pollTicker.C:
			agent.Poll()
			slog.Info("Poll completed")
		case <-reportTicker.C:
			if err := agent.Report(); err != nil {
				slog.Error("Report error", slog.String("error", err.Error()))
			} else {
				slog.Info("Report completed")
			}
		case <-ctx.Done():
			slog.Info("Agent shutting down...")
			return
		}
	}
}
