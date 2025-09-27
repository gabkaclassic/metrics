package main

import (
	"net/http"

	"github.com/gabkaclassic/metrics/internal/config"
	"github.com/gabkaclassic/metrics/internal/handler"
	"github.com/gabkaclassic/metrics/internal/repository"
	"github.com/gabkaclassic/metrics/internal/service"
	"github.com/gabkaclassic/metrics/internal/storage"
	"github.com/gabkaclassic/metrics/pkg/httpserver"
	"github.com/gabkaclassic/metrics/pkg/logger"
)

func main() {

	cfg := config.ParseServerConfig()

	logger.SetupLogger(logger.LogConfig(cfg.Log))

	router, err := setupRouter()

	panicWithError(err)

	server := httpserver.New(
		httpserver.Address(cfg.Address),
		httpserver.Handler(&router),
	)
	server.Run()
}

func setupRouter() (http.Handler, error) {

	storage := storage.NewMemStorage()

	// Metrics
	metricsRepository, err := repository.NewMetricsRepository(storage)

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

func panicWithError(err error) {
	if err != nil {
		panic(err)
	}
}
