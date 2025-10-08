package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gabkaclassic/metrics/internal/config"
	"github.com/gabkaclassic/metrics/internal/dump"
	"github.com/gabkaclassic/metrics/internal/handler"
	"github.com/gabkaclassic/metrics/internal/repository"
	"github.com/gabkaclassic/metrics/internal/service"
	"github.com/gabkaclassic/metrics/internal/storage"
	"github.com/gabkaclassic/metrics/pkg/httpserver"
	"github.com/gabkaclassic/metrics/pkg/logger"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := config.ParseServerConfig()

	if err != nil {
		return err
	}

	logger.SetupLogger(logger.LogConfig(cfg.Log))

	storage := storage.NewMemStorage()

	dumper, err := dump.NewDumper(cfg.Dump.FileStoragePath, storage)

	if err != nil {
		return err
	}

	readDump(cfg.Dump, dumper)

	router, err := setupRouter(storage)

	if err != nil {
		return err
	}

	server := httpserver.New(
		httpserver.Address(cfg.Address),
		httpserver.Handler(&router),
	)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go startDumper(ctx, dumper, cfg.Dump)

	go func() {
		slog.Info("Starting HTTP server...")
		if err := server.Run(); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP server error", slog.String("error", err.Error()))
			stop()
		}
	}()

	<-ctx.Done()
	slog.Info("Shutting down gracefully...")
	slog.Info("Shutdown complete")

	return nil
}

func readDump(cfg config.Dump, dumper *dump.Dumper) {
	if cfg.Restore {
		dumper.Read()
	}
}

func startDumper(ctx context.Context, dumper *dump.Dumper, cfg config.Dump) {
	ticker := time.NewTicker(cfg.StoreInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := dumper.Dump(); err != nil {
				slog.Error("Dump error", slog.String("error", err.Error()))
			} else {
				slog.Info("Dump completed")
			}
		case <-ctx.Done():
			slog.Info("Dumper stopped")
			return
		}
	}
}

func setupRouter(strg *storage.MemStorage) (http.Handler, error) {

	// Metrics
	metricsRepository, err := repository.NewMetricsRepository(strg)

	if err != nil {
		return nil, err
	}

	metricsService, err := service.NewMetricsService(metricsRepository)

	if err != nil {
		return nil, err
	}

	metricsHandler, err := handler.NewMetricsHandler(metricsService)

	if err != nil {
		return nil, err
	}

	return handler.SetupRouter(&handler.RouterConfiguration{
		MetricsHandler: metricsHandler,
	}), nil
}
