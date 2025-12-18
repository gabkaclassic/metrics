package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/gabkaclassic/metrics/internal/audit"
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
		return fmt.Errorf("failed to parse server configuration: %w", err)
	}
	logger.SetupLogger(logger.LogConfig(cfg.Log))

	var metricsRepository repository.MetricsRepository
	var dumper *dump.Dumper
	var dumperEnabled bool

	if len(cfg.DB.DSN) > 0 {
		storage, err := storage.NewDBStorage(cfg.DB)
		if err != nil {
			return fmt.Errorf("failed to initialize database storage: %w", err)
		}
		defer storage.Close()

		metricsRepository, err = repository.NewDBMetricsRepository(storage)
		if err != nil {
			return fmt.Errorf("failed to create metrics repository (DB): %w", err)
		}

		slog.Info("Using database storage")
	} else {
		storage := storage.NewMemStorage()
		storageMutex := &sync.RWMutex{}

		metricsRepository, err = repository.NewMemoryMetricsRepository(storage, storageMutex)
		if err != nil {
			return fmt.Errorf("failed to create metrics repository (in-memory): %w", err)
		}

		dumper, err = dump.NewDumper(cfg.Dump.FileStoragePath, metricsRepository)
		if err != nil {
			return fmt.Errorf("failed to initialize dumper: %w", err)
		}

		slog.Info("Using file storage with dumper", "dump_file", cfg.Dump.FileStoragePath)
		dumperEnabled = true
		defer dumper.Close()

		readDump(cfg.Dump, dumper)
	}

	auditor, err := audit.NewAudior(cfg.Audit)

	if err != nil {
		return fmt.Errorf("failed to create auditor: %w", err)
	}

	router, err := setupRouter(&metricsRepository, cfg.SignKey, auditor)
	if err != nil {
		return fmt.Errorf("failed to setup HTTP router: %w", err)
	}

	server := httpserver.New(
		httpserver.Address(cfg.Address),
		httpserver.Handler(&router),
	)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if dumperEnabled {
		go dumper.StartDumper(ctx, cfg.Dump)
		slog.Info("Dumper started")
	}

	go server.Run(ctx, stop)

	<-ctx.Done()
	slog.Info("Shutdown complete")

	return nil
}

func readDump(cfg config.Dump, dumper *dump.Dumper) {
	if cfg.Restore {
		dumper.Read()
	}
}

func setupRouter(metricsRepository *repository.MetricsRepository, signKey string, auditor audit.Auditor) (http.Handler, error) {

	// Metrics
	metricsService, err := service.NewMetricsService(*metricsRepository, auditor)

	if err != nil {
		return nil, err
	}

	metricsHandler, err := handler.NewMetricsHandler(metricsService)

	if err != nil {
		return nil, err
	}

	return handler.SetupRouter(&handler.RouterConfiguration{
		MetricsHandler: metricsHandler,
		SignKey:        signKey,
	}), nil
}
