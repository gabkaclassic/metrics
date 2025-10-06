package middleware

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gabkaclassic/metrics/pkg/compress"
	api "github.com/gabkaclassic/metrics/pkg/error"
	"github.com/google/uuid"
)

type middleware func(handler http.Handler) http.Handler

type ContentType string

const (
	JSON      ContentType = "application/json"
	TEXT      ContentType = "text/plain; charset=utf-8"
	HTML      ContentType = "text/html"
	HTML_UTF8 ContentType = "text/html; charset=utf-8"
)

type CompressType string

const (
	GZIP CompressType = "gzip"
)

var compressors = map[CompressType]func(http.ResponseWriter) (*compress.CompressWriter, error){
	GZIP: compress.NewGzipWriter,
}

var decompressors = map[CompressType]func(io.ReadCloser) (*compress.CompressReader, error){
	GZIP: compress.NewGzipReader,
}

func Compress(compressMapping map[ContentType]CompressType) middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			compressType, exists := compressMapping[ContentType(w.Header().Get("Content-Type"))]
			acceptedCompressTypes := r.Header.Get("Accept-Encoding")
			ctor, compressorExists := compressors[compressType]

			if !exists || !strings.Contains(acceptedCompressTypes, string(compressType)) || !compressorExists {
				next.ServeHTTP(w, r)
				return
			}

			writer, err := ctor(w)

			if err != nil {
				err := api.Internal("Create compressor failed", err)
				api.RespondError(w, err)
				return
			}

			defer writer.Close()

			writer.Header().Set("Content-Encoding", string(compressType))
			next.ServeHTTP(writer, r)
		})
	}
}

func Decompress() middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			compressType := r.Header.Get("Content-Encoding")
			ctor, decompressorExists := decompressors[CompressType(compressType)]

			if !decompressorExists {
				next.ServeHTTP(w, r)
				return
			}

			reader, err := ctor(r.Body)

			if err != nil {
				err := api.Internal("Create decompressor failed", err)
				api.RespondError(w, err)
				return
			}

			r.Body = reader

			next.ServeHTTP(w, r)
		})
	}
}

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

func RequireContentType(ct ContentType) middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Content-Type") != string(ct) {
				err := api.BadRequest("Invalid content type")
				api.RespondError(w, err)
				return
			}
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
