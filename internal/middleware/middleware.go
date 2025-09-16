package middleware

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type middleware func(handler http.Handler) http.Handler

func Wrap(handler http.Handler, middlewares ...middleware) http.Handler {
	wrapped := handler
	for _, handler := range middlewares {
		wrapped = handler(wrapped)
	}

	return wrapped
}

func TextPlainContentType(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")

		handler.ServeHTTP(w, r)
	})
}

func JSONContentType(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "application/json")

		handler.ServeHTTP(w, r)
	})
}

func Logger(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.NewString()
		}

		w.Header().Set("X-Request-ID", requestID)

		var bodyBytes []byte
		if r.Body != nil {
			bodyBytes, _ = io.ReadAll(r.Body)
		}
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		slog.Info("Incoming request",
			slog.String("id", requestID),
			slog.String("method", r.Method),
			slog.String("url", r.URL.String()),
			slog.Any("headers", r.Header),
			slog.String("body", string(bodyBytes)),
			slog.String("remote_addr", r.RemoteAddr),
		)

		start := time.Now()

		handler.ServeHTTP(w, r)

		duration := time.Since(start)
		slog.Info("Request processed",
			slog.String("id", requestID),
			slog.Duration("duration", duration),
		)
	})
}
