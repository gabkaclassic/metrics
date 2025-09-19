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

type ContentType string

const (
	JSON ContentType = "application/json"
	TEXT ContentType = "text/plain; charset=utf-8"
	HTML ContentType = "text/html; charset=utf-8"
)

func Wrap(h http.Handler, middlewares ...middleware) http.HandlerFunc {
	for _, m := range middlewares {
		h = m(h)
	}
	return h.ServeHTTP
}

func WithContentType(ct ContentType) middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", string(ct))
			next.ServeHTTP(w, r)
		})
	}
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
