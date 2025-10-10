package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/gabkaclassic/metrics/pkg/middleware"
)

type RouterConfiguration struct {
	MetricsHandler *MetricsHandler
}

func SetupRouter(config *RouterConfiguration) http.Handler {

	router := chi.NewRouter()

	router.Use(
		middleware.Logger,
		middleware.Decompress(),
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
			middleware.Compress(map[middleware.ContentType]middleware.CompressType{
				middleware.HTML:     middleware.GZIP,
				middleware.HTMLUTF8: middleware.GZIP,
			}),
			middleware.WithContentType(middleware.HTML),
		),
	)
	router.Post(
		"/update/",
		middleware.Wrap(
			http.HandlerFunc(handler.SaveJSON),
			middleware.RequireContentType(middleware.JSON),
			middleware.Compress(map[middleware.ContentType]middleware.CompressType{
				middleware.JSON: middleware.GZIP,
			}),
			middleware.WithContentType(middleware.JSON),
		),
	)
	router.Post(
		"/value/",
		middleware.Wrap(
			http.HandlerFunc(handler.GetJSON),
			middleware.RequireContentType(middleware.JSON),
			middleware.Compress(map[middleware.ContentType]middleware.CompressType{
				middleware.JSON: middleware.GZIP,
			}),
			middleware.WithContentType(middleware.JSON),
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
			middleware.Compress(map[middleware.ContentType]middleware.CompressType{
				middleware.JSON: middleware.GZIP,
			}),
			middleware.WithContentType(middleware.JSON),
		),
	)
}
