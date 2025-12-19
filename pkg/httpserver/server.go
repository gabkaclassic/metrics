package httpserver

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

const (
	// defaultAddress is used when no explicit server address is provided.
	defaultAddress = "0.0.0.0:8000"
)

// Server represents an HTTP server instance.
type Server struct {
	address string
	handler *http.Handler
}

// New creates a new Server instance configured with provided options.
//
// If no address is specified, defaultAddress is used.
func New(options ...Option) *Server {

	server := &Server{
		address: defaultAddress,
	}

	for _, option := range options {
		option(server)
	}

	return server
}

// GetHandler returns the configured HTTP handler.
func (server *Server) GetHandler() *http.Handler {
	return server.handler
}

// Run starts the HTTP server and blocks until context cancellation.
//
// The server listens on the configured address and performs a graceful
// shutdown with a fixed timeout after context cancellation.
// The provided stop function is called if shutdown fails.
func (server *Server) Run(ctx context.Context, stop context.CancelFunc) {
	slog.Info("Starting HTTP server...", slog.String("address", server.address))

	srv := &http.Server{
		Addr:    server.address,
		Handler: *server.handler,
	}

	errChan := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
		close(errChan)
	}()

	select {
	case err := <-errChan:
		if err != nil {
			slog.Error("HTTP server run error", slog.String("error", err.Error()))
			stop()
		}
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		stop()
		slog.Error("HTTP server shutdown error", slog.String("error", err.Error()))
	}

	slog.Info("HTTP server stopped gracefully")
}
