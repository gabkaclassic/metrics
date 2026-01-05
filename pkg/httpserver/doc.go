// Package httpserver provides a minimal HTTP server wrapper with
// graceful shutdown support.
//
// The package encapsulates net/http.Server initialization,
// configuration via functional options and controlled shutdown
// using context cancellation.
package httpserver
