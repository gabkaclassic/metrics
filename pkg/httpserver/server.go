package httpserver

import (
	"net/http"
)

const (
	defaultAddress = "0.0.0.0:8000"
)

type Server struct {
	address string
	handler http.Handler
}

func New(options ...Option) *Server {

	server := &Server{
		address: defaultAddress,
		handler: http.DefaultServeMux,
	}

	for _, option := range options {
		option(server)
	}

	return server
}

func (server *Server) Run() {
	err := http.ListenAndServe(server.address, server.handler)

	if err != nil {
		panic(err)
	}
}
