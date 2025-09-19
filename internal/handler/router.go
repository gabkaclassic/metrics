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
	router.Get(
		"/",
		middleware.Wrap(
			http.HandlerFunc(handler.GetAll),
			middleware.WithContentType(middleware.HTML),
		),
	)
	router.Post(
		"/update/{type}/{id}/{value}",
		middleware.Wrap(
			http.HandlerFunc(handler.Save),
			middleware.WithContentType(middleware.TEXT),
		),
	)
	router.Get(
		"/value/{type}/{id}",
		middleware.Wrap(
			http.HandlerFunc(handler.Get),
			middleware.WithContentType(middleware.JSON),
		),
	)
}
