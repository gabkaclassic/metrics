package httpclient

import (
	"io"
	"net/http"
	"net/url"
	"time"
)

type (
	// ResponseFilter defines a function to determine if a request should be retried.
	//
	// Returns true if the request should be retried, false otherwise.
	ResponseFilter func(resp *http.Response, err error) bool

	// RequestOptions holds optional parameters for an HTTP request.
	RequestOptions struct {
		Params  *Params   // URL query parameters
		Headers *Headers  // Request headers
		Body    io.Reader // Request body
	}

	// ResponseDelay returns a duration to wait before the next retry attempt.
	ResponseDelay func() time.Duration

	// DelayGenerator returns a ResponseDelay for a given retry attempt.
	DelayGenerator func(attempt int) ResponseDelay

	// Method represents HTTP methods as strings.
	Method string

	// Headers defines HTTP headers map.
	Headers map[string]string

	// Params defines URL query parameters map.
	Params map[string]string

	// HTTPClient defines the interface for an HTTP client supporting
	// standard HTTP methods.
	HTTPClient interface {
		Get(url string, opts *RequestOptions) (*http.Response, error)
		Post(url string, opts *RequestOptions) (*http.Response, error)
		Put(url string, opts *RequestOptions) (*http.Response, error)
		Patch(url string, opts *RequestOptions) (*http.Response, error)
		Delete(url string, opts *RequestOptions) (*http.Response, error)
	}

	// Client implements HTTPClient with retry logic, delays, headers,
	// base URL and response filtering.
	Client struct {
		baseURL        string
		responseFilter ResponseFilter
		maxRetries     int
		delay          DelayGenerator
		client         http.Client
		timeout        time.Duration
		headers        Headers
	}
)

// NewClient creates a new Client configured with functional options.
//
// Default behavior:
//   - maxRetries = 3
//   - JSON headers empty
//   - Default responseFilter retries on errors or 5xx status codes
//   - Delay between retries = attempt * 1s
func NewClient(options ...Option) *Client {
	c := &Client{
		client:         http.Client{},
		maxRetries:     3,
		headers:        make(Headers),
		responseFilter: func(resp *http.Response, err error) bool { return err != nil || resp.StatusCode >= 500 },
		delay: func(attempt int) ResponseDelay {
			return func() time.Duration { return time.Duration(attempt) * time.Second }
		},
	}

	for _, opt := range options {
		opt(c)
	}

	return c
}

// buildURL constructs a full URL with query parameters.
func buildURL(base string, params Params) string {
	if len(params) == 0 {
		return base
	}

	u, _ := url.Parse(base)
	q := u.Query()
	for k, v := range params {
		q.Set(k, url.QueryEscape(v))
	}
	u.RawQuery = q.Encode()
	return u.String()
}

// do executes an HTTP request with retries, headers and optional body.
//
// Internal method used by Get, Post, Put, Patch, Delete.
func (c *Client) do(url string, method string, opts *RequestOptions) (*http.Response, error) {
	var params Params
	var headers Headers
	var body io.Reader

	if opts != nil {
		if opts.Params != nil {
			params = *opts.Params
		}
		if opts.Headers != nil {
			headers = *opts.Headers
		}
		body = opts.Body
	}

	fullURL := buildURL(url, params)
	retries := 0
	var resp *http.Response
	var err error

	for {
		req, reqErr := http.NewRequest(method, fullURL, body)
		if reqErr != nil {
			return nil, reqErr
		}

		for k, v := range c.headers {
			req.Header.Set(k, v)
		}
		for k, v := range headers {
			req.Header.Set(k, v)
		}

		resp, err = c.client.Do(req)

		if !c.responseFilter(resp, err) || retries >= c.maxRetries {
			break
		}

		retries++
		delayFn := c.delay(retries)
		time.Sleep(delayFn())
	}

	return resp, err
}

// Get performs an HTTP GET request.
func (c *Client) Get(url string, opts *RequestOptions) (*http.Response, error) {
	return c.do(c.baseURL+url, http.MethodGet, opts)
}

// Post performs an HTTP POST request.
func (c *Client) Post(url string, opts *RequestOptions) (*http.Response, error) {
	return c.do(c.baseURL+url, http.MethodPost, opts)
}

// Put performs an HTTP PUT request.
func (c *Client) Put(url string, opts *RequestOptions) (*http.Response, error) {
	return c.do(c.baseURL+url, http.MethodPut, opts)
}

// Patch performs an HTTP PATCH request.
func (c *Client) Patch(url string, opts *RequestOptions) (*http.Response, error) {
	return c.do(c.baseURL+url, http.MethodPatch, opts)
}

// Delete performs an HTTP DELETE request.
func (c *Client) Delete(url string, opts *RequestOptions) (*http.Response, error) {
	return c.do(c.baseURL+url, http.MethodDelete, opts)
}
