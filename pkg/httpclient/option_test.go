package httpclient

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestOptions(t *testing.T) {
	tests := []struct {
		name            string
		option          Option
		expectedURL     string
		expectedTimeout time.Duration
		expectedHeaders Headers
		expectedRetries int
		expectedFilter  ResponseFilter
		expectedDelay   DelayGenerator
	}{
		{
			name:        "BaseURL sets URL",
			option:      BaseURL("https://api.example.com"),
			expectedURL: "https://api.example.com",
		},
		{
			name:            "Timeout sets timeout",
			option:          Timeout(30 * time.Second),
			expectedTimeout: 30 * time.Second,
		},
		{
			name:            "HeadersOption sets headers",
			option:          HeadersOption(Headers{"X-API-Key": "secret"}),
			expectedHeaders: Headers{"X-API-Key": "secret"},
		},
		{
			name:            "MaxRetries sets retries",
			option:          MaxRetries(3),
			expectedRetries: 3,
		},
		{
			name:           "Filter sets response filter",
			option:         Filter(func(resp *http.Response, err error) bool { return false }),
			expectedFilter: func(resp *http.Response, err error) bool { return false },
		},
		{
			name:          "Delay sets delay generator",
			option:        Delay(func(attempt int) ResponseDelay { return func() time.Duration { return time.Second } }),
			expectedDelay: func(attempt int) ResponseDelay { return func() time.Duration { return time.Second } },
		},
		{
			name:        "BaseURL with empty string",
			option:      BaseURL(""),
			expectedURL: "",
		},
		{
			name:            "Timeout zero duration",
			option:          Timeout(0),
			expectedTimeout: 0,
		},
		{
			name:            "MaxRetries zero",
			option:          MaxRetries(0),
			expectedRetries: 0,
		},
		{
			name:            "MaxRetries negative",
			option:          MaxRetries(-1),
			expectedRetries: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{}

			tt.option(c)

			if tt.expectedURL != "" {
				assert.Equal(t, tt.expectedURL, c.baseURL)
			}
			if tt.expectedTimeout != 0 {
				assert.Equal(t, tt.expectedTimeout, c.timeout)
				assert.Equal(t, tt.expectedTimeout, c.client.Timeout)
			}
			if tt.expectedHeaders != nil {
				assert.Equal(t, tt.expectedHeaders, c.headers)
			}
			if tt.name == "MaxRetries sets retries" || tt.name == "MaxRetries zero" || tt.name == "MaxRetries negative" {
				assert.Equal(t, tt.expectedRetries, c.maxRetries)
			}
			if tt.expectedFilter != nil {
				assert.NotNil(t, c.responseFilter)
			}
			if tt.expectedDelay != nil {
				assert.NotNil(t, c.delay)
			}
		})
	}
}
