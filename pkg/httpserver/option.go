package httpserver

import (
	"net/http"
)

type Option func(*Server)

func Address(address string) Option {
	return func(server *Server) {
		server.address = address
	}
}

func Handler(handler *http.ServeMux) Option {
	return func(server *Server) {
		server.handler = handler
	}
}
