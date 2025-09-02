package handler

import (
	"net/http"

	"github.com/gabkaclassic/metrics/internal/middleware"
)

type RouterConfiguration struct {
	Mux            *http.ServeMux
	MetricsHandler *MetricsHandler
}

func SetupRouter(config *RouterConfiguration) {

	// Metrics
	config.Mux.Handle(
		"/update/{type}/{id}/{value}",
		middleware.Wrap(
			http.HandlerFunc(config.MetricsHandler.Save),
			middleware.Logger,
			middleware.TextPlainContentType,
		),
	)
	config.Mux.Handle(
		"/get/{id}",
		middleware.Wrap(
			http.HandlerFunc(config.MetricsHandler.Get),
			middleware.Logger,
			middleware.JSONContentType,
		),
	)
}
