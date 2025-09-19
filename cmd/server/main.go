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

	router := setupRouter()
	server := httpserver.New(
		httpserver.Address(cfg.Address),
		httpserver.Handler(&router),
	)
	server.Run()
}

func setupRouter() http.Handler {

	storage := storage.NewMemStorage()

	// Metrics
	metricsRepository := repository.NewMetricsRepository(storage)
	metricsService := service.NewMetricsService(metricsRepository)
	metricsHandler := handler.NewMetricsHandler(metricsService)

	return handler.SetupRouter(&handler.RouterConfiguration{
		MetricsHandler: metricsHandler,
	})
}
