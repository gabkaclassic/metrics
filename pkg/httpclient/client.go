package httpclient

import (
	"io"
	"net/http"
	"net/url"
	"time"
)

type ResponseFilter func(resp *http.Response, err error) bool

type RequestOptions struct {
	Params  *Params
	Headers *Headers
	Body    io.Reader
}

type ResponseDelay func() time.Duration

type DelayGenerator func(attempt int) ResponseDelay

type Method string

type Headers map[string]string
type Params map[string]string

type HttpClient interface {
	Get(url string, opts *RequestOptions) (*http.Response, error)
	Post(url string, opts *RequestOptions) (*http.Response, error)
	Put(url string, opts *RequestOptions) (*http.Response, error)
	Patch(url string, opts *RequestOptions) (*http.Response, error)
	Delete(url string, opts *RequestOptions) (*http.Response, error)
}

type Client struct {
	baseUrl        string
	responseFilter ResponseFilter
	maxRetries     int
	delay          DelayGenerator
	client         http.Client
	timeout        time.Duration
	headers        Headers
}

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

func (c *Client) Get(url string, opts *RequestOptions) (*http.Response, error) {
	return c.do(c.baseUrl+url, http.MethodGet, opts)
}

func (c *Client) Post(url string, opts *RequestOptions) (*http.Response, error) {
	return c.do(c.baseUrl+url, http.MethodPost, opts)
}

func (c *Client) Put(url string, opts *RequestOptions) (*http.Response, error) {
	return c.do(c.baseUrl+url, http.MethodPut, opts)
}

func (c *Client) Patch(url string, opts *RequestOptions) (*http.Response, error) {
	return c.do(c.baseUrl+url, http.MethodPatch, opts)
}

func (c *Client) Delete(url string, opts *RequestOptions) (*http.Response, error) {
	return c.do(c.baseUrl+url, http.MethodDelete, opts)
}
