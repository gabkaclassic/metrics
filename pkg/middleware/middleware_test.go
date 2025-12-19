package middleware

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gabkaclassic/metrics/pkg/compress"

	"github.com/stretchr/testify/assert"
)

func TestWrap(t *testing.T) {
	tests := []struct {
		name         string
		middlewares  []middleware
		expectHeader string
	}{
		{
			name:         "no middleware",
			middlewares:  nil,
			expectHeader: "",
		},
		{
			name: "single middleware",
			middlewares: []middleware{
				func(next http.Handler) http.Handler {
					return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.Header().Set("X-Test", "value")
						next.ServeHTTP(w, r)
					})
				},
			},
			expectHeader: "value",
		},
		{
			name: "multiple middleware",
			middlewares: []middleware{
				func(next http.Handler) http.Handler {
					return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.Header().Set("X-Test", "1")
						next.ServeHTTP(w, r)
					})
				},
				func(next http.Handler) http.Handler {
					return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.Header().Set("X-Two", "2")
						next.ServeHTTP(w, r)
					})
				},
			},
			expectHeader: "1",
		},
		{
			name:         "without middleware",
			middlewares:  []middleware{},
			expectHeader: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})
			wrapped := Wrap(handler, tt.middlewares...)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rr := httptest.NewRecorder()

			wrapped.ServeHTTP(rr, req)

			if tt.expectHeader != "" {
				assert.Equal(t, tt.expectHeader, rr.Header().Get("X-Test"))
			} else {
				assert.Equal(t, http.StatusOK, rr.Code)
			}
		})
	}
}

func TestRequireContentType(t *testing.T) {
	tests := []struct {
		name           string
		requiredType   ContentType
		requestType    string
		expectStatus   int
		expectErrorMsg string
		expectNextCall bool
	}{
		{
			name:           "valid content type passes",
			requiredType:   JSON,
			requestType:    "application/json",
			expectStatus:   http.StatusOK,
			expectNextCall: true,
		},
		{
			name:           "invalid content type returns error",
			requiredType:   JSON,
			requestType:    "text/plain",
			expectStatus:   http.StatusBadRequest,
			expectErrorMsg: "Invalid content type",
			expectNextCall: false,
		},
		{
			name:           "missing content type returns error",
			requiredType:   JSON,
			requestType:    "",
			expectStatus:   http.StatusBadRequest,
			expectErrorMsg: "Invalid content type",
			expectNextCall: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextCalled := false

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				w.WriteHeader(http.StatusOK)
			})

			mw := RequireContentType(tt.requiredType)(next)

			req := httptest.NewRequest(http.MethodPost, "/", nil)
			if tt.requestType != "" {
				req.Header.Set("Content-Type", tt.requestType)
			}

			rr := httptest.NewRecorder()
			mw.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectStatus, rr.Code)
			assert.Equal(t, tt.expectNextCall, nextCalled)

			if tt.expectErrorMsg != "" {
				assert.Contains(t, rr.Body.String(), tt.expectErrorMsg)
			}
		})
	}
}

func TestWithContentType(t *testing.T) {
	tests := []struct {
		name         string
		ct           ContentType
		expectedType string
	}{
		{
			name:         "set JSON content type",
			ct:           JSON,
			expectedType: "application/json",
		},
		{
			name:         "set TEXT content type",
			ct:           TEXT,
			expectedType: "text/plain; charset=utf-8",
		},
		{
			name:         "set HTML UTF-8 content type",
			ct:           HTMLUTF8,
			expectedType: "text/html; charset=utf-8",
		},
		{
			name:         "set HTML content type",
			ct:           HTML,
			expectedType: "text/html",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			wrapped := WithContentType(tt.ct)(handler)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rr := httptest.NewRecorder()

			wrapped.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedType, rr.Header().Get("Content-Type"))
		})
	}
}

func TestLogger(t *testing.T) {
	tests := []struct {
		name           string
		requestID      string
		body           string
		expectHeaderID string
		expectStatus   int
	}{
		{
			name:           "no request ID, empty body",
			requestID:      "",
			body:           "",
			expectHeaderID: "",
			expectStatus:   http.StatusOK,
		},
		{
			name:           "with request ID and body",
			requestID:      "test-id",
			body:           "hello",
			expectHeaderID: "test-id",
			expectStatus:   http.StatusAccepted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(func() int {
					if tt.name == "with request ID and body" {
						return http.StatusAccepted
					}
					return http.StatusOK
				}())
			})

			wrapped := Logger(handler)

			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(tt.body))
			if tt.requestID != "" {
				req.Header.Set("X-Request-ID", tt.requestID)
			}
			rr := httptest.NewRecorder()

			wrapped.ServeHTTP(rr, req)

			headerID := rr.Header().Get("X-Request-ID")
			if tt.expectHeaderID == "" {
				assert.NotEmpty(t, headerID)
			} else {
				assert.Equal(t, tt.expectHeaderID, headerID)
			}
			assert.Equal(t, tt.expectStatus, rr.Code)
		})
	}
}

func TestSignVerify(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		signKey        string
		requestSign    string
		requestBody    string
		expectStatus   int
		expectNextCall bool
	}{
		{
			name:           "POST with valid signature passes",
			method:         http.MethodPost,
			signKey:        "test-key",
			requestSign:    "IaKG/W/Z9SZ2AHxm0PiD20bQYVjCZtM/tTfCO8YY5Wc=",
			requestBody:    "test-data",
			expectStatus:   http.StatusOK,
			expectNextCall: true,
		},
		{
			name:           "POST with invalid signature returns error",
			method:         http.MethodPost,
			signKey:        "test-key",
			requestSign:    "invalid-signature",
			requestBody:    "test-data",
			expectStatus:   http.StatusBadRequest,
			expectNextCall: false,
		},
		{
			name:           "POST with wrong key returns error",
			method:         http.MethodPost,
			signKey:        "wrong-key",
			requestSign:    "pXNY6Vs2c0dM7sBsXW6bQ3X6WJPSqcbql1k7p3G0n/g=",
			requestBody:    "test-data",
			expectStatus:   http.StatusBadRequest,
			expectNextCall: false,
		},
		{
			name:           "POST with empty body and valid signature passes",
			method:         http.MethodPost,
			signKey:        "test-key",
			requestSign:    "JxHMI+mrG4qbwP6ZEjjakmcWJKnr2vHBq+wG5+mhT5s=",
			requestBody:    "",
			expectStatus:   http.StatusOK,
			expectNextCall: true,
		},
		{
			name:           "GET method skips verification",
			method:         http.MethodGet,
			signKey:        "test-key",
			requestSign:    "",
			requestBody:    "test-data",
			expectStatus:   http.StatusOK,
			expectNextCall: true,
		},
		{
			name:           "PUT method skips verification",
			method:         http.MethodPut,
			signKey:        "test-key",
			requestSign:    "",
			requestBody:    "test-data",
			expectStatus:   http.StatusOK,
			expectNextCall: true,
		},
		{
			name:           "POST with malformed base64 returns error",
			method:         http.MethodPost,
			signKey:        "test-key",
			requestSign:    "!!!malformed!!!",
			requestBody:    "test-data",
			expectStatus:   http.StatusBadRequest,
			expectNextCall: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextCalled := false

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				w.WriteHeader(http.StatusOK)
			})

			mw := SignVerify(tt.signKey)(next)

			req := httptest.NewRequest(tt.method, "/", strings.NewReader(tt.requestBody))
			if tt.requestSign != "" {
				req.Header.Set("Hash", tt.requestSign)
			}

			rr := httptest.NewRecorder()
			mw.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectStatus, rr.Code)
			assert.Equal(t, tt.expectNextCall, nextCalled)
		})
	}
}

func TestAuditContext(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		xff        string
		expectIP   string
	}{
		{
			name:       "uses RemoteAddr when no X-Forwarded-For",
			remoteAddr: "10.0.0.1:12345",
			expectIP:   "10.0.0.1:12345",
		},
		{
			name:       "uses first X-Forwarded-For value",
			remoteAddr: "10.0.0.1:12345",
			xff:        "192.168.1.1, 192.168.1.2",
			expectIP:   "192.168.1.1",
		},
		{
			name:       "trims spaces in X-Forwarded-For",
			remoteAddr: "10.0.0.1:12345",
			xff:        "  172.16.0.1  ",
			expectIP:   "172.16.0.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotIP string
			var gotTS int64

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotIP, _ = r.Context().Value(ctxIPKey).(string)
				gotTS, _ = r.Context().Value(ctxTSKey).(int64)
				w.WriteHeader(http.StatusOK)
			})

			mw := AuditContext(next)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = tt.remoteAddr
			if tt.xff != "" {
				req.Header.Set("X-Forwarded-For", tt.xff)
			}

			rr := httptest.NewRecorder()
			before := time.Now().Unix()

			mw.ServeHTTP(rr, req)

			after := time.Now().Unix()

			assert.Equal(t, http.StatusOK, rr.Code)
			assert.Equal(t, tt.expectIP, gotIP)
			assert.True(t, gotTS >= before && gotTS <= after)
		})
	}
}

func TestAuditIPFromCtx(t *testing.T) {
	tests := []struct {
		name     string
		ctx      context.Context
		expected string
	}{
		{
			name:     "ctx with valid IP string",
			ctx:      context.WithValue(context.Background(), ctxIPKey, "192.168.1.1"),
			expected: "192.168.1.1",
		},
		{
			name:     "ctx with wrong value type",
			ctx:      context.WithValue(context.Background(), ctxIPKey, 123),
			expected: "",
		},
		{
			name:     "ctx without IP key",
			ctx:      context.Background(),
			expected: "",
		},
		{
			name:     "ctx with nil value",
			ctx:      context.WithValue(context.Background(), ctxIPKey, nil),
			expected: "",
		},
		{
			name:     "ctx with empty string",
			ctx:      context.WithValue(context.Background(), ctxIPKey, ""),
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AuditIPFromCtx(tt.ctx)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAuditTsFromCtx(t *testing.T) {
	tests := []struct {
		name     string
		ctx      context.Context
		expected int64
	}{
		{
			name:     "ctx with valid timestamp",
			ctx:      context.WithValue(context.Background(), ctxTSKey, int64(1672531200)),
			expected: 1672531200,
		},
		{
			name:     "ctx with wrong value type",
			ctx:      context.WithValue(context.Background(), ctxTSKey, "not-a-timestamp"),
			expected: 0,
		},
		{
			name:     "ctx without TS key",
			ctx:      context.Background(),
			expected: 0,
		},
		{
			name:     "ctx with nil value",
			ctx:      context.WithValue(context.Background(), ctxTSKey, nil),
			expected: 0,
		},
		{
			name:     "ctx with zero value",
			ctx:      context.WithValue(context.Background(), ctxTSKey, int64(0)),
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AuditTsFromCtx(tt.ctx)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCompressMiddleware(t *testing.T) {
	tests := []struct {
		name             string
		contentType      string
		acceptEncoding   string
		compressMapping  map[ContentType]CompressType
		expectCompressed bool
		expectNextCall   bool
	}{
		{
			name:           "no compression when accept-encoding does not match",
			contentType:    "application/json",
			acceptEncoding: "br",
			compressMapping: map[ContentType]CompressType{
				"application/json": "gzip",
			},
			expectCompressed: false,
			expectNextCall:   true,
		},
		{
			name:           "no compression when content-type not in mapping",
			contentType:    "text/plain",
			acceptEncoding: "gzip",
			compressMapping: map[ContentType]CompressType{
				"application/json": "gzip",
			},
			expectCompressed: false,
			expectNextCall:   true,
		},
		{
			name:           "no compression when compressor does not exist",
			contentType:    "application/json",
			acceptEncoding: "gzip",
			compressMapping: map[ContentType]CompressType{
				"application/json": "gzip",
			},
			expectCompressed: false,
			expectNextCall:   true,
		},
	}

	originalCompressors := compressors
	defer func() { compressors = originalCompressors }()

	compressors = map[CompressType]func(http.ResponseWriter) (*compress.CompressWriter, error){
		GZIP: compress.NewGzipWriter,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextCalled := false

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				w.Header().Set("Content-Type", tt.contentType)
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("response-body"))
			})

			mw := Compress(tt.compressMapping)(next)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Accept-Encoding", tt.acceptEncoding)

			rr := httptest.NewRecorder()
			mw.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectNextCall, nextCalled)

			encoding := rr.Header().Get("Content-Encoding")
			if tt.expectCompressed {
				assert.Equal(t, "gzip", encoding)
			} else {
				assert.Empty(t, encoding)
			}
		})
	}
}

func TestDecompressMiddleware(t *testing.T) {
	tests := []struct {
		name            string
		contentEncoding string
		body            []byte
		decompressors   map[CompressType]func(io.ReadCloser) (io.ReadCloser, error)
		expectBody      string
		expectStatus    int
		expectNextCall  bool
	}{
		{
			name:            "gzip decompression applied",
			contentEncoding: "gzip",
			body: func() []byte {
				var buf bytes.Buffer
				w := gzip.NewWriter(&buf)
				w.Write([]byte("payload"))
				w.Close()
				return buf.Bytes()
			}(),
			decompressors: map[CompressType]func(io.ReadCloser) (io.ReadCloser, error){
				"gzip": func(r io.ReadCloser) (io.ReadCloser, error) {
					gr, err := gzip.NewReader(r)
					if err != nil {
						return nil, err
					}
					return gr, nil
				},
			},
			expectBody:     "payload",
			expectStatus:   http.StatusOK,
			expectNextCall: true,
		},
		{
			name:            "unknown encoding passes through",
			contentEncoding: "br",
			body:            []byte("raw"),
			decompressors:   map[CompressType]func(io.ReadCloser) (io.ReadCloser, error){},
			expectBody:      "raw",
			expectStatus:    http.StatusOK,
			expectNextCall:  true,
		},
		{
			name:            "decompressor ctor error returns 500",
			contentEncoding: "gzip",
			body:            []byte("broken"),
			decompressors: map[CompressType]func(io.ReadCloser) (io.ReadCloser, error){
				"gzip": func(r io.ReadCloser) (io.ReadCloser, error) {
					return nil, errors.New("boom")
				},
			},
			expectStatus:   http.StatusInternalServerError,
			expectNextCall: false,
		},
	}

	originalDecompressors := decompressors
	defer func() { decompressors = originalDecompressors }()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			nextCalled := false
			var receivedBody []byte

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				receivedBody, _ = io.ReadAll(r.Body)
				w.WriteHeader(http.StatusOK)
			})

			mw := Decompress()(next)

			req := httptest.NewRequest(http.MethodPost, "/", io.NopCloser(bytes.NewReader(tt.body)))
			if tt.contentEncoding != "" {
				req.Header.Set("Content-Encoding", tt.contentEncoding)
			}

			rr := httptest.NewRecorder()
			mw.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectStatus, rr.Code)
			assert.Equal(t, tt.expectNextCall, nextCalled)

			if tt.expectNextCall && tt.expectBody != "" {
				assert.Equal(t, tt.expectBody, string(receivedBody))
			}
		})
	}
}
