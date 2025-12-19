package handler

import (
	"net/http"

	"github.com/gabkaclassic/metrics/pkg/middleware"
	"github.com/go-chi/chi/v5"
)

type RouterConfiguration struct {
	MetricsHandler *MetricsHandler
	SignKey        string
}

func SetupRouter(config *RouterConfiguration) http.Handler {

	router := chi.NewRouter()

	router.Use(
		middleware.Logger,
		middleware.AuditContext,
	)

	// Ping endpoint
	router.Get("/ping", func(w http.ResponseWriter, r *http.Request) {})

	setupMetricsRouter(router, config.MetricsHandler, middleware.Decompress(), middleware.SignVerify(config.SignKey))

	return router
}

func setupMetricsRouter(
	router *chi.Mux,
	handler *MetricsHandler,
	decompressMiddleware func(handler http.Handler) http.Handler,
	signVerifyMiddleware func(handler http.Handler) http.Handler,
) {
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
			decompressMiddleware,
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
			decompressMiddleware,
			signVerifyMiddleware,
		),
	)
	router.Post(
		"/updates/",
		middleware.Wrap(
			http.HandlerFunc(handler.SaveAll),
			middleware.RequireContentType(middleware.JSON),
			middleware.Compress(map[middleware.ContentType]middleware.CompressType{
				middleware.JSON: middleware.GZIP,
			}),
			middleware.WithContentType(middleware.JSON),
			decompressMiddleware,
			signVerifyMiddleware,
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
			decompressMiddleware,
		),
	)
	router.Post(
		"/update/{type}/{id}/{value}",
		middleware.Wrap(
			http.HandlerFunc(handler.Save),
			middleware.WithContentType(middleware.TEXT),
			decompressMiddleware,
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
			decompressMiddleware,
		),
	)
}
