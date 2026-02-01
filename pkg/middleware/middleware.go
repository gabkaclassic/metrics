// Package middleware provides HTTP middleware utilities.
//
// The package contains composable middleware functions used to:
//   - Validate request integrity (signature verification)
//   - Compress and decompress HTTP bodies
//   - Enforce and set Content-Type headers
//   - Log incoming HTTP requests
//   - Inject audit metadata into request context
//
// Middlewares are designed to be combined using Wrap.
package middleware

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gabkaclassic/metrics/pkg/compress"
	"github.com/gabkaclassic/metrics/pkg/crypt"
	api "github.com/gabkaclassic/metrics/pkg/error"
	"github.com/gabkaclassic/metrics/pkg/hash"
	"github.com/google/uuid"
)

type (
	// middleware represents a standard HTTP middleware function.
	middleware func(handler http.Handler) http.Handler

	// ContentType represents an HTTP Content-Type value.
	ContentType string

	// ContextKey represents a typed key for context values.
	ContextKey string

	// CompressType represents a supported HTTP compression algorithm.
	CompressType string
)

const (
	// Supported content types.
	JSON     ContentType = "application/json"
	TEXT     ContentType = "text/plain; charset=utf-8"
	HTML     ContentType = "text/html"
	HTMLUTF8 ContentType = "text/html; charset=utf-8"

	// Supported compression types.
	GZIP CompressType = "gzip"

	// Context keys used for audit metadata.
	ctxIPKey ContextKey = "sourceIP"
	ctxTSKey ContextKey = "ts"
)

var compressors = map[CompressType]func(http.ResponseWriter) (*compress.CompressWriter, error){
	GZIP: compress.NewGzipWriter,
}

var decompressors = map[CompressType]func(io.ReadCloser) (*compress.CompressReader, error){
	GZIP: compress.NewGzipReader,
}

// SignVerify returns a middleware that verifies request body integrity.
//
// The middleware validates the request body using a SHA-256 HMAC signature
// provided in the "Hash" header.
//
// Applied only to POST requests without Accept-Encoding header.
// If signature verification fails, the request is rejected with 400 status.
func SignVerify(signKey string) middleware {

	verifier := hash.NewSHA256Verifier(signKey)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			if r.Method != http.MethodPost || r.Header.Get("Accept-Encoding") != "" {
				next.ServeHTTP(w, r)
				return
			}

			sign := r.Header.Get("Hash")
			var bodyBytes []byte
			var err error
			if r.Body != nil {
				bodyBytes, err = io.ReadAll(r.Body)

				if err != nil {
					apiErr := api.Internal("Internal server error", err)
					api.RespondError(w, apiErr)
					return
				}
			}
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			if !verifier.Verify(bodyBytes, sign) {
				err := api.BadRequest("Data sign is invalid")
				api.RespondError(w, err)
				return
			}
			slog.Debug("Data sign verified successful")

			next.ServeHTTP(w, r)
		})
	}
}

// Compress returns a middleware that compresses HTTP responses.
//
// Compression is applied when:
//   - Response Content-Type matches compressMapping
//   - Client declares support via Accept-Encoding header
//
// Currently supports gzip compression.
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

// Decompress returns a middleware that decompresses request bodies.
//
// If Content-Encoding header matches a supported compression type,
// the request body is transparently decompressed before being passed
// to the next handler.
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

func Decrypt(privateKeyPath string) (func(http.Handler) http.Handler, error) {

	var decryptor *crypt.Decryptor
	var err error
	if len(privateKeyPath) > 0 {
		decryptor, err = crypt.NewDecryptor(privateKeyPath)

		if err != nil {
			return nil, err
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if decryptor == nil || r.Body == nil {
				next.ServeHTTP(w, r)
				return
			}

			encrypted, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			decrypted, err := decryptor.Decrypt(encrypted)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			r.Body = io.NopCloser(bytes.NewReader(decrypted))
			r.ContentLength = int64(len(decrypted))

			next.ServeHTTP(w, r)
		})
	}, nil
}

// Wrap applies a chain of middlewares to an HTTP handler.
//
// Middlewares are applied in the order they are provided.
func Wrap(h http.Handler, middlewares ...middleware) http.HandlerFunc {
	for _, m := range middlewares {
		h = m(h)
	}
	return h.ServeHTTP
}

// WithContentType returns a middleware that sets the response Content-Type.
func WithContentType(ct ContentType) middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", string(ct))
			next.ServeHTTP(w, r)
		})
	}
}

// RequireContentType returns a middleware that enforces request Content-Type.
//
// Requests with a mismatched Content-Type are rejected with 400 status.
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

// Logger logs incoming HTTP requests and their processing time.
//
// The middleware logs:
//   - Request ID (generated if missing)
//   - HTTP method and URL
//   - Request headers and body
//   - Client address
//   - Request processing duration
func Logger(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.NewString()
		}

		w.Header().Set("X-Request-ID", requestID)

		var bodyBytes []byte
		var err error
		if r.Body != nil {
			bodyBytes, err = io.ReadAll(r.Body)

			if err != nil {
				apiErr := api.Internal("Internal server error", err)
				api.RespondError(w, apiErr)
				return
			}
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

// AuditContext injects audit metadata into the request context.
//
// The middleware extracts the client IP address and request timestamp
// and stores them in the request context for downstream consumers.
func AuditContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			ip = strings.TrimSpace(strings.Split(xff, ",")[0])
		}

		ctx := context.WithValue(r.Context(), ctxIPKey, ip)
		ctx = context.WithValue(ctx, ctxTSKey, time.Now().Unix())

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// AuditIPFromCtx extracts the source IP address from context.
//
// Returns empty string if the value is not present.
func AuditIPFromCtx(ctx context.Context) string {
	if v, ok := ctx.Value(ctxIPKey).(string); ok {
		return v
	}
	return ""
}

// AuditTSFromCtx extracts the audit timestamp from context.
//
// Returns zero if the value is not present.
func AuditTSFromCtx(ctx context.Context) int64 {
	if v, ok := ctx.Value(ctxTSKey).(int64); ok {
		return v
	}
	return 0
}
