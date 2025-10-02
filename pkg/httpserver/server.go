package httpserver

import (
	"log/slog"
	"net/http"
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

func (server *Server) Run() {
	slog.Info("Run HHTP server", slog.String("address", server.address))
	srv := http.Server{
		Addr:    server.address,
		Handler: *server.handler,
	}

	err := srv.ListenAndServe()

	if err != nil {
		panic(err)
	}
}
