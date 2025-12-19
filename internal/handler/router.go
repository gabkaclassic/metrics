package handler

import (
	"net/http"

	"github.com/gabkaclassic/metrics/pkg/middleware"
	"github.com/go-chi/chi/v5"
)

// RouterConfiguration holds the dependencies and settings required to setup the HTTP router.
// Used to decouple router configuration from application initialization.
type RouterConfiguration struct {
	// MetricsHandler handles all metrics-related HTTP endpoints.
	// Must be initialized before router setup.
	MetricsHandler *MetricsHandler

	// SignKey is the secret key used for request signature verification.
	// If empty, signature verification middleware is disabled.
	SignKey string
}

// SetupRouter configures and returns a fully initialized HTTP router with all middleware.
//
// config: Router configuration containing handler and security settings.
//
// Returns:
//   - http.Handler: Configured router with all middleware applied.
//
// The router includes:
//   - Request logging
//   - Audit context propagation
//   - Compression/decompression
//   - Content type validation
//   - Request signature verification (if SignKey provided)
//
// Routes configured:
//   - GET  /ping     - Health check endpoint
//   - GET  /         - HTML metrics dashboard
//   - POST /update/  - JSON metric update (single)
//   - POST /updates/ - JSON metric batch update
//   - POST /value/   - JSON metric retrieval
//   - POST /update/{type}/{id}/{value} - Plain text metric update
//   - GET  /value/{type}/{id} - Plain text metric retrieval
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

// setupMetricsRouter configures all metrics-related routes with appropriate middleware.
// This function separates metrics route configuration for better organization.
//
// router: Chi router instance to register routes on.
// handler: Metrics handler implementing endpoint logic.
// decompressMiddleware: Middleware for decompressing request bodies (gzip).
// signVerifyMiddleware: Middleware for verifying request signatures (HMAC).
//
// Middleware composition per route:
//   - All routes: decompression, content type headers
//   - Write operations: signature verification (if key provided)
//   - JSON endpoints: content type validation, compression
//   - HTML endpoint: HTML-specific compression
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
