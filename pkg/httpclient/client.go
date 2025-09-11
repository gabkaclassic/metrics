package httpclient

import (
	"io"
	"net/http"
	"net/url"
	"time"
)

type ResponseFilter func(resp *http.Response, err error) bool

type ResponseDelay func() time.Duration

type DelayGenerator func(attempt int) ResponseDelay

type Method string

type Headers map[string]string
type Params map[string]string

type HttpClient interface {
	Get(url string, params Params) (<-chan *http.Response, <-chan error)
	Post(url string, body io.Reader) (<-chan *http.Response, <-chan error)
	Put(url string, body io.Reader) (<-chan *http.Response, <-chan error)
	Patch(url string, body io.Reader) (<-chan *http.Response, <-chan error)
	Delete(url string, params Params, body io.Reader) (<-chan *http.Response, <-chan error)
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

func (c *Client) do(url string, method string, headers Headers, body io.Reader) (*http.Response, error) {
	retries := 0
	var resp *http.Response
	var err error

	for {
		req, reqErr := http.NewRequest(method, url, body)
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

func (c *Client) asyncCall(url string, method string, headers Headers, body io.Reader) (<-chan *http.Response, <-chan error) {
	respCh := make(chan *http.Response, 1)
	errCh := make(chan error, 1)

	go func() {
		resp, err := c.do(url, method, headers, body)
		if err != nil {
			errCh <- err
			close(respCh)
			close(errCh)
			return
		}
		respCh <- resp
		close(respCh)
		close(errCh)
	}()

	return respCh, errCh
}

func buildURL(base string, params Params) string {
	if len(params) == 0 {
		return base
	}

	u, _ := url.Parse(base)
	q := u.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()
	return u.String()
}

func (c *Client) Get(url string, params Params) (<-chan *http.Response, <-chan error) {
	fullURL := buildURL(c.baseUrl+url, params)
	return c.asyncCall(fullURL, http.MethodGet, nil, nil)
}

func (c *Client) Post(url string, body io.Reader) (<-chan *http.Response, <-chan error) {
	return c.asyncCall(c.baseUrl+url, http.MethodPost, nil, body)
}

func (c *Client) Put(url string, body io.Reader) (<-chan *http.Response, <-chan error) {
	return c.asyncCall(c.baseUrl+url, http.MethodPut, nil, body)
}

func (c *Client) Patch(url string, body io.Reader) (<-chan *http.Response, <-chan error) {
	return c.asyncCall(c.baseUrl+url, http.MethodPatch, nil, body)
}

func (c *Client) Delete(url string, params Params, body io.Reader) (<-chan *http.Response, <-chan error) {
	fullURL := buildURL(c.baseUrl+url, params)
	return c.asyncCall(fullURL, http.MethodDelete, nil, body)
}
