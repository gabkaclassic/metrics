package httpserver

import (
	"net/http"
)

const (
	defaultAddress = "0.0.0.0:8000"
)

type Server struct {
	address string
	handler *http.ServeMux
}

func New(options ...Option) *Server {

	server := &Server{
		address: defaultAddress,
		handler: http.NewServeMux(),
	}

	for _, option := range options {
		option(server)
	}

	return server
}

func (server *Server) GetHandler() *http.ServeMux {
	return server.handler
}

func (server *Server) Run() {
	err := http.ListenAndServe(server.address, server.handler)

	if err != nil {
		panic(err)
	}
}
