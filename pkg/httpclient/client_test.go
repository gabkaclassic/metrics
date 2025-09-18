package httpclient

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
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
				func(c *Client) { c.baseUrl = "http://example.com" },
			},
			check: func(t *testing.T, c *Client) {
				assert.Equal(t, 10, c.maxRetries)
				assert.Equal(t, 5*time.Second, c.timeout)
				assert.Equal(t, "http://example.com", c.baseUrl)
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

func TestClient_do(t *testing.T) {
	tests := []struct {
		name          string
		serverHandler http.HandlerFunc
		clientHeaders Headers
		callHeaders   Headers
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
			clientHeaders: Headers{"X-Test1": "val1"},
			callHeaders:   Headers{"X-Test2": "val2"},
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
				headers:        tt.clientHeaders,
				maxRetries:     tt.maxRetries,
				responseFilter: tt.respFilter,
				delay: func(attempt int) ResponseDelay {
					return func() time.Duration { return 0 }
				},
			}

			resp, err := c.do(testURL, http.MethodGet, tt.callHeaders, io.NopCloser(bytes.NewBufferString("body")))
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}

func TestClient_asyncCall(t *testing.T) {
	tests := []struct {
		name          string
		serverHandler http.HandlerFunc
		expectErr     bool
		expectStatus  int
	}{
		{
			name: "successful request",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("ok"))
			},
			expectErr:    false,
			expectStatus: http.StatusOK,
		},
		{
			name: "server error",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			expectErr:    false,
			expectStatus: http.StatusInternalServerError,
		},
		{
			name:          "invalid url",
			serverHandler: nil,
			expectErr:     true,
			expectStatus:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var testURL string
			if tt.name == "invalid url" {
				testURL = "http://[::1]:badport"
			} else {
				srv := httptest.NewServer(tt.serverHandler)
				defer srv.Close()
				testURL = srv.URL
			}

			c := &Client{
				client:         http.Client{Timeout: time.Second},
				headers:        Headers{},
				maxRetries:     0,
				responseFilter: func(resp *http.Response, err error) bool { return false },
				delay:          func(int) ResponseDelay { return func() time.Duration { return 0 } },
			}

			respCh, errCh := c.asyncCall(testURL, http.MethodGet, nil, io.NopCloser(bytes.NewBufferString("body")))

			select {
			case err := <-errCh:
				if tt.expectErr {
					assert.Error(t, err)
					assert.Nil(t, <-respCh)
				} else {
					assert.NoError(t, err)
				}
			case resp := <-respCh:
				if !tt.expectErr {
					assert.NotNil(t, resp)
					assert.Equal(t, tt.expectStatus, resp.StatusCode)
				}
			case <-time.After(2 * time.Second):
				t.Fatal("timeout waiting for asyncCall")
			}
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
			want:   "http://example.com/api?a=1&x=9",
		},
		{
			name:   "special characters",
			base:   "http://example.com/api",
			params: Params{"q": "hello world"},
			want:   "http://example.com/api?q=hello+world",
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

func TestClient_Get(t *testing.T) {
	tests := []struct {
		name         string
		baseURL      string
		relativePath string
		params       Params
		handler      http.HandlerFunc
		expectErr    bool
		expectStatus int
	}{
		{
			name:         "simple get no params",
			baseURL:      "",
			relativePath: "/ok",
			params:       nil,
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
			params:       Params{"q": "go", "page": "2"},
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
				baseUrl:        base,
				client:         http.Client{Timeout: time.Second},
				headers:        Headers{},
				maxRetries:     0,
				responseFilter: func(resp *http.Response, err error) bool { return false },
				delay:          func(int) ResponseDelay { return func() time.Duration { return 0 } },
			}

			respCh, errCh := c.Get(tt.relativePath, tt.params)

			select {
			case err := <-errCh:
				if tt.expectErr {
					assert.Error(t, err)
					assert.Nil(t, <-respCh)
				} else {
					assert.NoError(t, err)
				}
			case resp := <-respCh:
				if !tt.expectErr {
					assert.NotNil(t, resp)
					assert.Equal(t, tt.expectStatus, resp.StatusCode)
				}
			case <-time.After(2 * time.Second):
				t.Fatal("timeout waiting for Get")
			}
		})
	}
}

func TestClient_Post(t *testing.T) {
	tests := []struct {
		name         string
		baseURL      string
		relativePath string
		body         io.Reader
		handler      http.HandlerFunc
		expectErr    bool
		expectStatus int
		expectBody   string
	}{
		{
			name:         "successful post with body",
			baseURL:      "",
			relativePath: "/submit",
			body:         bytes.NewBufferString("hello"),
			handler: func(w http.ResponseWriter, r *http.Request) {
				data, _ := io.ReadAll(r.Body)
				assert.Equal(t, "hello", string(data))
				assert.Equal(t, http.MethodPost, r.Method)
				w.WriteHeader(http.StatusCreated)
			},
			expectErr:    false,
			expectStatus: http.StatusCreated,
			expectBody:   "hello",
		},
		{
			name:         "post without body",
			baseURL:      "",
			relativePath: "/empty",
			body:         nil,
			handler: func(w http.ResponseWriter, r *http.Request) {
				data, _ := io.ReadAll(r.Body)
				assert.Equal(t, "", string(data))
				assert.Equal(t, http.MethodPost, r.Method)
				w.WriteHeader(http.StatusOK)
			},
			expectErr:    false,
			expectStatus: http.StatusOK,
			expectBody:   "",
		},
		{
			name:         "invalid url",
			baseURL:      "http://[::1]:badport",
			relativePath: "/fail",
			body:         bytes.NewBufferString("fail"),
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
				baseUrl:        base,
				client:         http.Client{Timeout: time.Second},
				headers:        Headers{},
				maxRetries:     0,
				responseFilter: func(resp *http.Response, err error) bool { return false },
				delay:          func(int) ResponseDelay { return func() time.Duration { return 0 } },
			}

			respCh, errCh := c.Post(tt.relativePath, tt.body)

			select {
			case err := <-errCh:
				if tt.expectErr {
					assert.Error(t, err)
					assert.Nil(t, <-respCh)
				} else {
					assert.NoError(t, err)
				}
			case resp := <-respCh:
				if !tt.expectErr {
					assert.NotNil(t, resp)
					assert.Equal(t, tt.expectStatus, resp.StatusCode)
				}
			case <-time.After(2 * time.Second):
				t.Fatal("timeout waiting for Post")
			}
		})
	}
}

func TestClient_Put(t *testing.T) {
	tests := []struct {
		name         string
		baseURL      string
		relativePath string
		body         io.Reader
		handler      http.HandlerFunc
		expectErr    bool
		expectStatus int
	}{
		{
			name:         "successful put with body",
			baseURL:      "",
			relativePath: "/update",
			body:         bytes.NewBufferString("data123"),
			handler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPut, r.Method)
				b, _ := io.ReadAll(r.Body)
				assert.Equal(t, "data123", string(b))
				w.WriteHeader(http.StatusOK)
			},
			expectErr:    false,
			expectStatus: http.StatusOK,
		},
		{
			name:         "put without body",
			baseURL:      "",
			relativePath: "/nobody",
			body:         nil,
			handler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPut, r.Method)
				b, _ := io.ReadAll(r.Body)
				assert.Equal(t, "", string(b))
				w.WriteHeader(http.StatusNoContent)
			},
			expectErr:    false,
			expectStatus: http.StatusNoContent,
		},
		{
			name:         "invalid url",
			baseURL:      "http://[::1]:badport",
			relativePath: "/fail",
			body:         bytes.NewBufferString("fail"),
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
				baseUrl:        base,
				client:         http.Client{Timeout: time.Second},
				headers:        Headers{},
				maxRetries:     0,
				responseFilter: func(resp *http.Response, err error) bool { return false },
				delay:          func(int) ResponseDelay { return func() time.Duration { return 0 } },
			}

			respCh, errCh := c.Put(tt.relativePath, tt.body)

			select {
			case err := <-errCh:
				if tt.expectErr {
					assert.Error(t, err)
					assert.Nil(t, <-respCh)
				} else {
					assert.NoError(t, err)
				}
			case resp := <-respCh:
				if !tt.expectErr {
					assert.NotNil(t, resp)
					assert.Equal(t, tt.expectStatus, resp.StatusCode)
				}
			case <-time.After(2 * time.Second):
				t.Fatal("timeout waiting for Put")
			}
		})
	}
}

func TestClient_Patch(t *testing.T) {
	tests := []struct {
		name         string
		baseURL      string
		relativePath string
		body         io.Reader
		handler      http.HandlerFunc
		expectErr    bool
		expectStatus int
	}{
		{
			name:         "successful patch with body",
			baseURL:      "",
			relativePath: "/modify",
			body:         bytes.NewBufferString("patchdata"),
			handler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPatch, r.Method)
				b, _ := io.ReadAll(r.Body)
				assert.Equal(t, "patchdata", string(b))
				w.WriteHeader(http.StatusOK)
			},
			expectErr:    false,
			expectStatus: http.StatusOK,
		},
		{
			name:         "patch without body",
			baseURL:      "",
			relativePath: "/nobody",
			body:         nil,
			handler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPatch, r.Method)
				b, _ := io.ReadAll(r.Body)
				assert.Equal(t, "", string(b))
				w.WriteHeader(http.StatusNoContent)
			},
			expectErr:    false,
			expectStatus: http.StatusNoContent,
		},
		{
			name:         "invalid url",
			baseURL:      "http://[::1]:badport",
			relativePath: "/fail",
			body:         bytes.NewBufferString("fail"),
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
				baseUrl:        base,
				client:         http.Client{Timeout: time.Second},
				headers:        Headers{},
				maxRetries:     0,
				responseFilter: func(resp *http.Response, err error) bool { return false },
				delay:          func(int) ResponseDelay { return func() time.Duration { return 0 } },
			}

			respCh, errCh := c.Patch(tt.relativePath, tt.body)

			select {
			case err := <-errCh:
				if tt.expectErr {
					assert.Error(t, err)
					assert.Nil(t, <-respCh)
				} else {
					assert.NoError(t, err)
				}
			case resp := <-respCh:
				if !tt.expectErr {
					assert.NotNil(t, resp)
					assert.Equal(t, tt.expectStatus, resp.StatusCode)
				}
			case <-time.After(2 * time.Second):
				t.Fatal("timeout waiting for Patch")
			}
		})
	}
}

func TestClient_Delete(t *testing.T) {
	tests := []struct {
		name         string
		baseURL      string
		relativePath string
		params       Params
		body         io.Reader
		handler      http.HandlerFunc
		expectErr    bool
		expectStatus int
	}{
		{
			name:         "successful delete with body and params",
			baseURL:      "",
			relativePath: "/remove",
			params:       Params{"id": "123"},
			body:         bytes.NewBufferString("delete-me"),
			handler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodDelete, r.Method)
				assert.Equal(t, "123", r.URL.Query().Get("id"))
				b, _ := io.ReadAll(r.Body)
				assert.Equal(t, "delete-me", string(b))
				w.WriteHeader(http.StatusOK)
			},
			expectErr:    false,
			expectStatus: http.StatusOK,
		},
		{
			name:         "delete without body",
			baseURL:      "",
			relativePath: "/nobody",
			params:       Params{"force": "true"},
			body:         nil,
			handler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodDelete, r.Method)
				assert.Equal(t, "true", r.URL.Query().Get("force"))
				b, _ := io.ReadAll(r.Body)
				assert.Equal(t, "", string(b))
				w.WriteHeader(http.StatusNoContent)
			},
			expectErr:    false,
			expectStatus: http.StatusNoContent,
		},
		{
			name:         "invalid url",
			baseURL:      "http://[::1]:badport",
			relativePath: "/fail",
			params:       nil,
			body:         bytes.NewBufferString("fail"),
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
				baseUrl:        base,
				client:         http.Client{Timeout: time.Second},
				headers:        Headers{},
				maxRetries:     0,
				responseFilter: func(resp *http.Response, err error) bool { return false },
				delay:          func(int) ResponseDelay { return func() time.Duration { return 0 } },
			}

			respCh, errCh := c.Delete(tt.relativePath, tt.params, tt.body)

			select {
			case err := <-errCh:
				if tt.expectErr {
					assert.Error(t, err)
					assert.Nil(t, <-respCh)
				} else {
					assert.NoError(t, err)
				}
			case resp := <-respCh:
				if !tt.expectErr {
					assert.NotNil(t, resp)
					assert.Equal(t, tt.expectStatus, resp.StatusCode)
				}
			case <-time.After(2 * time.Second):
				t.Fatal("timeout waiting for Delete")
			}
		})
	}
}
