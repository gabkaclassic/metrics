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
		middleware.TextPlainContentType(http.HandlerFunc(config.MetricsHandler.Save)),
	)
	config.Mux.Handle(
		"/get/{id}",
		middleware.JSONContentType(http.HandlerFunc(config.MetricsHandler.Get)),
	)
}
