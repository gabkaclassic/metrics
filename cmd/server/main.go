package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

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
	storageMutex := &sync.RWMutex{}

	dumper, err := dump.NewDumper(cfg.Dump.FileStoragePath, storage, storageMutex)

	if err != nil {
		return err
	}

	defer dumper.Close()

	readDump(cfg.Dump, dumper)

	router, err := setupRouter(storage, storageMutex)

	if err != nil {
		return err
	}

	server := httpserver.New(
		httpserver.Address(cfg.Address),
		httpserver.Handler(&router),
	)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go dumper.StartDumper(ctx, cfg.Dump)
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

func setupRouter(strg *storage.MemStorage, storageMutex *sync.RWMutex) (http.Handler, error) {

	// Metrics
	metricsRepository, err := repository.NewMetricsRepository(strg, storageMutex)

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
