package httpserver

import (
	"net/http"
)

// Option represents a functional option for Server configuration.
type Option func(*Server)

// Address sets the server listen address.
func Address(address string) Option {
	return func(server *Server) {
		server.address = address
	}
}

// Handler sets the HTTP handler for the server.
func Handler(handler *http.Handler) Option {
	return func(server *Server) {
		server.handler = handler
	}
}
