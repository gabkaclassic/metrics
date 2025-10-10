package httpserver

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

const (
	defaultAddress = "0.0.0.0:8000"
)

type Server struct {
	address string
	handler *http.Handler
}

func New(options ...Option) *Server {

	server := &Server{
		address: defaultAddress,
	}

	for _, option := range options {
		option(server)
	}

	return server
}

func (server *Server) GetHandler() *http.Handler {
	return server.handler
}

func (server *Server) Run(ctx context.Context, stop context.CancelFunc) {
	slog.Info("Starting HTTP server...", slog.String("address", server.address))

	srv := &http.Server{
		Addr:    server.address,
		Handler: *server.handler,
	}

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("HTTP server run error", slog.String("error", err.Error()))
	}

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {

		stop()
		slog.Error("HTTP server error", slog.String("error", err.Error()))
	}

	slog.Info("HTTP server stopped gracefully")
}
