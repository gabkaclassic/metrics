package main

import (
	"github.com/gabkaclassic/metrics/internal/config"
	"github.com/gabkaclassic/metrics/internal/handler"
	"github.com/gabkaclassic/metrics/internal/repository"
	"github.com/gabkaclassic/metrics/internal/service"
	"github.com/gabkaclassic/metrics/internal/storage"
	"github.com/gabkaclassic/metrics/pkg/httpserver"
)

func main() {

	cfg := config.ParseServerConfig()

	server := httpserver.New(
		httpserver.Address(cfg.Address),
	)

	setupRouter(server)

	server.Run()
}

func setupRouter(server *httpserver.Server) {

	storage := storage.NewMemStorage()

	// Metrics
	metricsRepository := repository.NewMetricsRepository(storage)
	metricsService := service.NewMetricsService(metricsRepository)
	metricsHandler := handler.NewMetricsHandler(metricsService)

	handler.SetupRouter(&handler.RouterConfiguration{
		Mux:            server.GetHandler(),
		MetricsHandler: metricsHandler,
	})
}
