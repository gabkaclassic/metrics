package handler

import (
	"github.com/go-chi/chi/v5"
	"net/http"

	"github.com/gabkaclassic/metrics/internal/middleware"
)

type RouterConfiguration struct {
	MetricsHandler *MetricsHandler
}

func SetupRouter(config *RouterConfiguration) http.Handler {

	router := chi.NewRouter()

	router.Use(
		middleware.Logger,
	)

	setupMetricsRouter(router, config.MetricsHandler)

	return router
}

func setupMetricsRouter(router *chi.Mux, handler *MetricsHandler) {
	// Metrics
	router.Handle(
		"/update/{type}/{id}/{value}",
		middleware.Wrap(
			http.HandlerFunc(handler.Save),
			middleware.TextPlainContentType,
		),
	)
	router.Handle(
		"/get/{id}",
		middleware.Wrap(
			http.HandlerFunc(handler.Get),
			middleware.JSONContentType,
		),
	)
}
