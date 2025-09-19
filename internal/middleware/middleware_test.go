package middleware

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
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
			name:         "set HTML content type",
			ct:           HTML,
			expectedType: "text/html; charset=utf-8",
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
