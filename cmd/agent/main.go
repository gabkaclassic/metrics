package main

import (
	"log/slog"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/gabkaclassic/metrics/internal/agent"
	"github.com/gabkaclassic/metrics/internal/config"
	"github.com/gabkaclassic/metrics/pkg/httpclient"
)

func main() {
	cfg := config.ParseAgentConfig()

	client := httpclient.NewClient(
		httpclient.BaseURL(cfg.Client.BaseUrl),
		httpclient.Timeout(cfg.Client.Timeout),
		httpclient.MaxRetries(cfg.Client.Retries),
	)

	agent := agent.NewAgent(
		client, &runtime.MemStats{},
	)

	pollTicker := time.NewTicker(cfg.PollInterval)
	reportTicker := time.NewTicker(cfg.ReportInterval)
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
