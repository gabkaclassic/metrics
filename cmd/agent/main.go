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
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	printTags()

	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func printTags() {
	if buildVersion == "" {
		buildVersion = "N/A"
	}
	if buildDate == "" {
		buildDate = "N/A"
	}
	if buildCommit == "" {
		buildCommit = "N/A"
	}

	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)
}

func run() error {
	cfg, err := config.ParseAgentConfig()
	if err != nil {
		return fmt.Errorf("failed to parse agent configuration: %w", err)
	}

	logger.SetupLogger(logger.LogConfig(cfg.Log))

	var client *httpclient.Client = nil
	if len(cfg.Client.BaseURL) > 0 {
		client = httpclient.NewClient(
			httpclient.BaseURL(cfg.Client.BaseURL),
			httpclient.Timeout(cfg.Client.Timeout),
			httpclient.MaxRetries(cfg.Client.Retries),
		)
	}

	var grpcConnection *grpc.ClientConn = nil
	if len(cfg.GRPCAddress) > 0 {
		grpcConnection, err = grpc.NewClient(cfg.GRPCAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			slog.Info("GRPC connection error", "error", err)
			grpcConnection = nil
		}
		defer grpcConnection.Close()
	}

	agent, err := agent.NewAgent(
		client, grpcConnection, cfg.BatchesEnabled, cfg.SignKey, cfg.PublicKeyPath, cfg.RateLimit, cfg.BatchSize,
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
