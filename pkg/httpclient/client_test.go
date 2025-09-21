package httpclient

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		options []Option
		check   func(t *testing.T, c *Client)
	}{
		{
			name:    "default client",
			options: nil,
			check: func(t *testing.T, c *Client) {
				assert.Equal(t, 3, c.maxRetries)
				assert.NotNil(t, c.headers)
				assert.NotNil(t, c.responseFilter)
				assert.NotNil(t, c.delay)
				assert.Equal(t, http.Client{}, c.client)

				resp := &http.Response{StatusCode: 500}
				assert.True(t, c.responseFilter(resp, nil))

				delayFn := c.delay(2)
				assert.Equal(t, 2*time.Second, delayFn())
			},
		},
		{
			name: "custom options applied",
			options: []Option{
				func(c *Client) { c.maxRetries = 10 },
				func(c *Client) { c.timeout = 5 * time.Second },
				func(c *Client) { c.baseURL = "http://example.com" },
			},
			check: func(t *testing.T, c *Client) {
				assert.Equal(t, 10, c.maxRetries)
				assert.Equal(t, 5*time.Second, c.timeout)
				assert.Equal(t, "http://example.com", c.baseURL)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.options...)
			tt.check(t, client)
		})
	}
}

func TestBuildURL(t *testing.T) {
	tests := []struct {
		name   string
		base   string
		params Params
		want   string
	}{
		{
			name:   "no params",
			base:   "http://example.com/api",
			params: Params{},
			want:   "http://example.com/api",
		},
		{
			name:   "single param",
			base:   "http://example.com/api",
			params: Params{"key": "value"},
			want:   "http://example.com/api?key=value",
		},
		{
			name:   "multiple params",
			base:   "http://example.com/api",
			params: Params{"a": "1", "b": "2"},
			want:   "http://example.com/api?a=1&b=2",
		},
		{
			name:   "base with existing query",
			base:   "http://example.com/api?x=9",
			params: Params{"a": "1"},
			want:   "http://example.com/api?x=9&a=1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildURL(tt.base, tt.params)

			wantURL, _ := url.Parse(tt.want)
			gotURL, _ := url.Parse(got)

			assert.Equal(t, wantURL.Scheme, gotURL.Scheme)
			assert.Equal(t, wantURL.Host, gotURL.Host)
			assert.Equal(t, wantURL.Path, gotURL.Path)
			assert.Equal(t, wantURL.Query(), gotURL.Query())
		})
	}
}

func TestClient_do(t *testing.T) {
	tests := []struct {
		name          string
		serverHandler http.HandlerFunc
		clientHeaders *Headers
		callHeaders   *Headers
		params        *Params
		maxRetries    int
		respFilter    ResponseFilter
		wantStatus    int
		wantErr       bool
	}{
		{
			name: "successful request no retries",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "val1", r.Header.Get("X-Test1"))
				assert.Equal(t, "val2", r.Header.Get("X-Test2"))
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("ok"))
			},
			clientHeaders: &Headers{"X-Test1": "val1"},
			callHeaders:   &Headers{"X-Test2": "val2"},
			params:        nil,
			maxRetries:    3,
			respFilter:    func(resp *http.Response, err error) bool { return false },
			wantStatus:    http.StatusOK,
			wantErr:       false,
		},
		{
			name: "request with retry then success",
			serverHandler: func() http.HandlerFunc {
				count := 0
				return func(w http.ResponseWriter, r *http.Request) {
					if count == 0 {
						count++
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					w.WriteHeader(http.StatusCreated)
				}
			}(),
			clientHeaders: nil,
			callHeaders:   nil,
			params:        nil,
			maxRetries:    2,
			respFilter:    func(resp *http.Response, err error) bool { return resp.StatusCode >= 500 },
			wantStatus:    http.StatusCreated,
			wantErr:       false,
		},
		{
			name: "request fails after max retries",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			clientHeaders: nil,
			callHeaders:   nil,
			params:        nil,
			maxRetries:    1,
			respFilter:    func(resp *http.Response, err error) bool { return resp.StatusCode >= 500 },
			wantStatus:    http.StatusInternalServerError,
			wantErr:       false,
		},
		{
			name: "invalid request url",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			clientHeaders: nil,
			callHeaders:   nil,
			params:        nil,
			maxRetries:    1,
			respFilter:    func(resp *http.Response, err error) bool { return false },
			wantStatus:    0,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var testURL string
			if tt.name != "invalid request url" {
				srv := httptest.NewServer(tt.serverHandler)
				defer srv.Close()
				testURL = srv.URL
			} else {
				testURL = "http://[::1]:namedport"
			}

			c := &Client{
				client:         http.Client{},
				headers:        nilOrMap(tt.clientHeaders),
				maxRetries:     tt.maxRetries,
				responseFilter: tt.respFilter,
				delay: func(attempt int) ResponseDelay {
					return func() time.Duration { return 0 }
				},
			}

			opts := &RequestOptions{
				Params:  tt.params,
				Headers: tt.callHeaders,
				Body:    nil,
			}

			resp, err := c.do(testURL, http.MethodGet, opts)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
			defer resp.Body.Close()
		})
	}
}

func TestClient_Get(t *testing.T) {
	tests := []struct {
		name         string
		baseURL      string
		relativePath string
		params       *Params
		headers      *Headers
		handler      http.HandlerFunc
		expectErr    bool
		expectStatus int
	}{
		{
			name:         "simple get no params",
			baseURL:      "",
			relativePath: "/ok",
			params:       nil,
			headers:      nil,
			handler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/ok", r.URL.Path)
				w.WriteHeader(http.StatusOK)
			},
			expectErr:    false,
			expectStatus: http.StatusOK,
		},
		{
			name:         "get with query params",
			baseURL:      "",
			relativePath: "/search",
			params:       &Params{"q": "go", "page": "2"},
			headers:      nil,
			handler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "go", r.URL.Query().Get("q"))
				assert.Equal(t, "2", r.URL.Query().Get("page"))
				w.WriteHeader(http.StatusAccepted)
			},
			expectErr:    false,
			expectStatus: http.StatusAccepted,
		},
		{
			name:         "invalid url",
			baseURL:      "http://[::1]:badport",
			relativePath: "/fail",
			params:       nil,
			headers:      nil,
			handler:      nil,
			expectErr:    true,
			expectStatus: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var base string
			if tt.handler != nil {
				srv := httptest.NewServer(tt.handler)
				defer srv.Close()
				base = srv.URL
			} else {
				base = tt.baseURL
			}

			c := &Client{
				baseURL:        base,
				client:         http.Client{Timeout: time.Second},
				headers:        nil,
				maxRetries:     0,
				responseFilter: func(resp *http.Response, err error) bool { return false },
				delay:          func(int) ResponseDelay { return func() time.Duration { return 0 } },
			}

			opts := &RequestOptions{
				Params:  tt.params,
				Headers: tt.headers,
				Body:    nil,
			}

			resp, err := c.Get(tt.relativePath, opts)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				defer resp.Body.Close()
				assert.Equal(t, tt.expectStatus, resp.StatusCode)
			}
		})
	}
}

func TestClient_Post(t *testing.T) {
	tests := []struct {
		name         string
		handler      http.HandlerFunc
		body         string
		params       *Params
		headers      *Headers
		maxRetries   int
		respFilter   ResponseFilter
		expectErr    bool
		expectStatus int
	}{
		{
			name: "simple post success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				data, _ := io.ReadAll(r.Body)
				assert.Equal(t, "body1", string(data))
				w.WriteHeader(http.StatusCreated)
			},
			body:         "body1",
			params:       nil,
			headers:      nil,
			maxRetries:   0,
			respFilter:   func(resp *http.Response, err error) bool { return false },
			expectErr:    false,
			expectStatus: http.StatusCreated,
		},
		{
			name: "post with headers",
			handler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "val1", r.Header.Get("X-Test1"))
				w.WriteHeader(http.StatusOK)
			},
			body:         "body2",
			params:       nil,
			headers:      &Headers{"X-Test1": "val1"},
			maxRetries:   0,
			respFilter:   func(resp *http.Response, err error) bool { return false },
			expectErr:    false,
			expectStatus: http.StatusOK,
		},
		{
			name: "post with query params",
			handler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "123", r.URL.Query().Get("id"))
				w.WriteHeader(http.StatusAccepted)
			},
			body:         "body3",
			params:       &Params{"id": "123"},
			headers:      nil,
			maxRetries:   0,
			respFilter:   func(resp *http.Response, err error) bool { return false },
			expectErr:    false,
			expectStatus: http.StatusAccepted,
		},
		{
			name: "post with retry then success",
			handler: func() http.HandlerFunc {
				count := 0
				return func(w http.ResponseWriter, r *http.Request) {
					if count == 0 {
						count++
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					w.WriteHeader(http.StatusOK)
				}
			}(),
			body:         "body4",
			params:       nil,
			headers:      nil,
			maxRetries:   2,
			respFilter:   func(resp *http.Response, err error) bool { return resp.StatusCode >= 500 },
			expectErr:    false,
			expectStatus: http.StatusOK,
		},
		{
			name:         "invalid URL",
			handler:      nil,
			body:         "body5",
			params:       nil,
			headers:      nil,
			maxRetries:   0,
			respFilter:   func(resp *http.Response, err error) bool { return false },
			expectErr:    true,
			expectStatus: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var base string
			if tt.handler != nil {
				srv := httptest.NewServer(tt.handler)
				defer srv.Close()
				base = srv.URL
			} else {
				base = "http://[::1]:badport"
			}

			c := &Client{
				baseURL:        base,
				client:         http.Client{Timeout: time.Second},
				headers:        nil,
				maxRetries:     tt.maxRetries,
				responseFilter: tt.respFilter,
				delay:          func(int) ResponseDelay { return func() time.Duration { return 0 } },
			}

			opts := &RequestOptions{
				Params:  tt.params,
				Headers: tt.headers,
				Body:    bytes.NewBufferString(tt.body),
			}

			resp, err := c.Post("/", opts)
			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				defer resp.Body.Close()
				assert.Equal(t, tt.expectStatus, resp.StatusCode)
			}
		})
	}
}

func TestClient_Put(t *testing.T) {
	tests := []struct {
		name         string
		handler      http.HandlerFunc
		body         string
		params       *Params
		headers      *Headers
		maxRetries   int
		respFilter   ResponseFilter
		expectErr    bool
		expectStatus int
	}{
		{
			name: "simple put success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				data, _ := io.ReadAll(r.Body)
				assert.Equal(t, "put body1", string(data))
				w.WriteHeader(http.StatusOK)
			},
			body:         "put body1",
			params:       nil,
			headers:      nil,
			maxRetries:   0,
			respFilter:   func(resp *http.Response, err error) bool { return false },
			expectErr:    false,
			expectStatus: http.StatusOK,
		},
		{
			name: "put with headers",
			handler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "val1", r.Header.Get("X-Test1"))
				w.WriteHeader(http.StatusAccepted)
			},
			body:         "put body2",
			params:       nil,
			headers:      &Headers{"X-Test1": "val1"},
			maxRetries:   0,
			respFilter:   func(resp *http.Response, err error) bool { return false },
			expectErr:    false,
			expectStatus: http.StatusAccepted,
		},
		{
			name: "put with query params",
			handler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "999", r.URL.Query().Get("id"))
				w.WriteHeader(http.StatusCreated)
			},
			body:         "put body3",
			params:       &Params{"id": "999"},
			headers:      nil,
			maxRetries:   0,
			respFilter:   func(resp *http.Response, err error) bool { return false },
			expectErr:    false,
			expectStatus: http.StatusCreated,
		},
		{
			name: "put with retry then success",
			handler: func() http.HandlerFunc {
				count := 0
				return func(w http.ResponseWriter, r *http.Request) {
					if count == 0 {
						count++
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					w.WriteHeader(http.StatusOK)
				}
			}(),
			body:         "put body4",
			params:       nil,
			headers:      nil,
			maxRetries:   2,
			respFilter:   func(resp *http.Response, err error) bool { return resp.StatusCode >= 500 },
			expectErr:    false,
			expectStatus: http.StatusOK,
		},
		{
			name:         "invalid URL",
			handler:      nil,
			body:         "put body5",
			params:       nil,
			headers:      nil,
			maxRetries:   0,
			respFilter:   func(resp *http.Response, err error) bool { return false },
			expectErr:    true,
			expectStatus: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var base string
			if tt.handler != nil {
				srv := httptest.NewServer(tt.handler)
				defer srv.Close()
				base = srv.URL
			} else {
				base = "http://[::1]:badport"
			}

			c := &Client{
				baseURL:        base,
				client:         http.Client{Timeout: time.Second},
				headers:        nil,
				maxRetries:     tt.maxRetries,
				responseFilter: tt.respFilter,
				delay:          func(int) ResponseDelay { return func() time.Duration { return 0 } },
			}

			opts := &RequestOptions{
				Params:  tt.params,
				Headers: tt.headers,
				Body:    bytes.NewBufferString(tt.body),
			}

			resp, err := c.Put("/", opts)
			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				defer resp.Body.Close()
				assert.Equal(t, tt.expectStatus, resp.StatusCode)
			}
		})
	}
}

func TestClient_Patch(t *testing.T) {
	tests := []struct {
		name         string
		handler      http.HandlerFunc
		body         string
		params       *Params
		headers      *Headers
		maxRetries   int
		respFilter   ResponseFilter
		expectErr    bool
		expectStatus int
	}{
		{
			name: "simple patch success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				data, _ := io.ReadAll(r.Body)
				assert.Equal(t, "patch body1", string(data))
				w.WriteHeader(http.StatusOK)
			},
			body:         "patch body1",
			params:       nil,
			headers:      nil,
			maxRetries:   0,
			respFilter:   func(resp *http.Response, err error) bool { return false },
			expectErr:    false,
			expectStatus: http.StatusOK,
		},
		{
			name: "patch with headers",
			handler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "val1", r.Header.Get("X-Test1"))
				w.WriteHeader(http.StatusAccepted)
			},
			body:         "patch body2",
			params:       nil,
			headers:      &Headers{"X-Test1": "val1"},
			maxRetries:   0,
			respFilter:   func(resp *http.Response, err error) bool { return false },
			expectErr:    false,
			expectStatus: http.StatusAccepted,
		},
		{
			name: "patch with query params",
			handler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "999", r.URL.Query().Get("id"))
				w.WriteHeader(http.StatusCreated)
			},
			body:         "patch body3",
			params:       &Params{"id": "999"},
			headers:      nil,
			maxRetries:   0,
			respFilter:   func(resp *http.Response, err error) bool { return false },
			expectErr:    false,
			expectStatus: http.StatusCreated,
		},
		{
			name: "patch with retry then success",
			handler: func() http.HandlerFunc {
				count := 0
				return func(w http.ResponseWriter, r *http.Request) {
					if count == 0 {
						count++
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					w.WriteHeader(http.StatusOK)
				}
			}(),
			body:         "patch body4",
			params:       nil,
			headers:      nil,
			maxRetries:   2,
			respFilter:   func(resp *http.Response, err error) bool { return resp.StatusCode >= 500 },
			expectErr:    false,
			expectStatus: http.StatusOK,
		},
		{
			name:         "invalid URL",
			handler:      nil,
			body:         "patch body5",
			params:       nil,
			headers:      nil,
			maxRetries:   0,
			respFilter:   func(resp *http.Response, err error) bool { return false },
			expectErr:    true,
			expectStatus: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var base string
			if tt.handler != nil {
				srv := httptest.NewServer(tt.handler)
				defer srv.Close()
				base = srv.URL
			} else {
				base = "http://[::1]:badport"
			}

			c := &Client{
				baseURL:        base,
				client:         http.Client{Timeout: time.Second},
				headers:        nil,
				maxRetries:     tt.maxRetries,
				responseFilter: tt.respFilter,
				delay:          func(int) ResponseDelay { return func() time.Duration { return 0 } },
			}

			opts := &RequestOptions{
				Params:  tt.params,
				Headers: tt.headers,
				Body:    bytes.NewBufferString(tt.body),
			}

			resp, err := c.Patch("/", opts)
			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				defer resp.Body.Close()
				assert.Equal(t, tt.expectStatus, resp.StatusCode)
			}
		})
	}
}

func TestClient_Delete(t *testing.T) {
	tests := []struct {
		name         string
		handler      http.HandlerFunc
		body         string
		params       *Params
		headers      *Headers
		maxRetries   int
		respFilter   ResponseFilter
		expectErr    bool
		expectStatus int
	}{
		{
			name: "simple delete success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			body:         "",
			params:       nil,
			headers:      nil,
			maxRetries:   0,
			respFilter:   func(resp *http.Response, err error) bool { return false },
			expectErr:    false,
			expectStatus: http.StatusOK,
		},
		{
			name: "delete with headers",
			handler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "val1", r.Header.Get("X-Test1"))
				w.WriteHeader(http.StatusAccepted)
			},
			body:         "",
			params:       nil,
			headers:      &Headers{"X-Test1": "val1"},
			maxRetries:   0,
			respFilter:   func(resp *http.Response, err error) bool { return false },
			expectErr:    false,
			expectStatus: http.StatusAccepted,
		},
		{
			name: "delete with query params",
			handler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "123", r.URL.Query().Get("id"))
				w.WriteHeader(http.StatusNoContent)
			},
			body:         "",
			params:       &Params{"id": "123"},
			headers:      nil,
			maxRetries:   0,
			respFilter:   func(resp *http.Response, err error) bool { return false },
			expectErr:    false,
			expectStatus: http.StatusNoContent,
		},
		{
			name: "delete with retry then success",
			handler: func() http.HandlerFunc {
				count := 0
				return func(w http.ResponseWriter, r *http.Request) {
					if count == 0 {
						count++
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					w.WriteHeader(http.StatusOK)
				}
			}(),
			body:         "",
			params:       nil,
			headers:      nil,
			maxRetries:   2,
			respFilter:   func(resp *http.Response, err error) bool { return resp.StatusCode >= 500 },
			expectErr:    false,
			expectStatus: http.StatusOK,
		},
		{
			name:         "invalid URL",
			handler:      nil,
			body:         "",
			params:       nil,
			headers:      nil,
			maxRetries:   0,
			respFilter:   func(resp *http.Response, err error) bool { return false },
			expectErr:    true,
			expectStatus: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var base string
			if tt.handler != nil {
				srv := httptest.NewServer(tt.handler)
				defer srv.Close()
				base = srv.URL
			} else {
				base = "http://[::1]:badport"
			}

			c := &Client{
				baseURL:        base,
				client:         http.Client{Timeout: time.Second},
				headers:        nil,
				maxRetries:     tt.maxRetries,
				responseFilter: tt.respFilter,
				delay:          func(int) ResponseDelay { return func() time.Duration { return 0 } },
			}

			opts := &RequestOptions{
				Params:  tt.params,
				Headers: tt.headers,
				Body:    bytes.NewBufferString(tt.body),
			}

			resp, err := c.Delete("/", opts)
			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				defer resp.Body.Close()
				assert.Equal(t, tt.expectStatus, resp.StatusCode)
			}
		})
	}
}

func nilOrMap(h *Headers) Headers {
	if h != nil {
		return *h
	}
	return nil
}
