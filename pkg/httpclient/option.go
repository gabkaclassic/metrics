package httpclient

import "time"

type Option func(*Client)

func BaseURL(url string) Option {
	return func(c *Client) {
		c.baseUrl = url
	}
}

func Timeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.timeout = timeout
		c.client.Timeout = timeout
	}
}

func HeadersOption(headers Headers) Option {
	return func(c *Client) {
		c.headers = headers
	}
}

func MaxRetries(n int) Option {
	return func(c *Client) {
		c.maxRetries = n
	}
}

func Filter(filter ResponseFilter) Option {
	return func(c *Client) {
		c.responseFilter = filter
	}
}

func Delay(generator DelayGenerator) Option {
	return func(c *Client) {
		c.delay = generator
	}
}
