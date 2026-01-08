package httpclient

import "time"

// Option delete performs an HTTP DELETE request.
type Option func(*Client)

// BaseURL sets the base URL for all requests.
func BaseURL(url string) Option {
	return func(c *Client) {
		c.baseURL = url
	}
}

// Timeout sets the HTTP client timeout.
func Timeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.timeout = timeout
		c.client.Timeout = timeout
	}
}

// HeadersOption sets default headers for all requests.
func HeadersOption(headers Headers) Option {
	return func(c *Client) {
		c.headers = headers
	}
}

// MaxRetries sets maximum retry attempts.
func MaxRetries(n int) Option {
	return func(c *Client) {
		c.maxRetries = n
	}
}

// Filter sets the response filter function to decide retries.
func Filter(filter ResponseFilter) Option {
	return func(c *Client) {
		c.responseFilter = filter
	}
}

// Delay sets the retry delay generator function.
func Delay(generator DelayGenerator) Option {
	return func(c *Client) {
		c.delay = generator
	}
}
