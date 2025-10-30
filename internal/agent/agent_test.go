package agent

import (
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/gabkaclassic/metrics/internal/model"
	"github.com/gabkaclassic/metrics/pkg/httpclient"
	"github.com/gabkaclassic/metrics/pkg/metric"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewAgent(t *testing.T) {
	dummyClient := httpclient.NewMockHTTPClient(t)
	stats := &runtime.MemStats{}

	agent := NewAgent(dummyClient, stats, true)

	assert.NotNil(t, agent)
	assert.Equal(t, dummyClient, agent.client)
	assert.Equal(t, stats, agent.stats)
	assert.NotNil(t, agent.metrics)
	assert.GreaterOrEqual(t, len(agent.metrics), 2)

	foundPollCount := false
	foundRandomValue := false
	for _, m := range agent.metrics {
		switch m.(type) {
		case *metric.PollCount:
			foundPollCount = true
		case *metric.RandomValue:
			foundRandomValue = true
		}
	}
	assert.True(t, foundPollCount)
	assert.True(t, foundRandomValue)
}

func TestMetricsAgent_Poll(t *testing.T) {
	stats := &runtime.MemStats{}

	m1 := metric.NewMockMetric(t)
	m2 := metric.NewMockMetric(t)

	m1.EXPECT().Update().Return()
	m2.EXPECT().Update().Return()

	agent := &MetricsAgent{
		stats: stats,
		metrics: []metric.Metric{
			m1,
			m2,
		},
	}

	agent.Poll()

	assert.NotZero(t, stats.Alloc)
}

func TestMetricsAgent_sendRequest(t *testing.T) {
	tests := []struct {
		name           string
		mockResponse   *http.Response
		mockError      error
		expectedErrMsg string
	}{
		{
			name: "success status 200",
			mockResponse: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"status":"ok"}`)),
			},
			mockError:      nil,
			expectedErrMsg: "",
		},
		{
			name:           "error http client returns error",
			mockResponse:   nil,
			mockError:      errors.New("network error"),
			expectedErrMsg: "send request error: network error",
		},
		{
			name: "error server returns 500",
			mockResponse: &http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       io.NopCloser(strings.NewReader(`{"error":"server failed"}`)),
			},
			mockError:      nil,
			expectedErrMsg: "request failed with status 500: {\"error\":\"server failed\"}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := httpclient.NewMockHTTPClient(t)
			endpoint := "http://example.com/api"
			body := bytes.NewBufferString("test body")

			mockClient.EXPECT().
				Post(endpoint, mock.AnythingOfType("*httpclient.RequestOptions")).
				Return(tt.mockResponse, tt.mockError)

			m := &MetricsAgent{
				client: mockClient,
				mu:     &sync.RWMutex{},
			}

			err := m.sendRequest(endpoint, body)

			if tt.expectedErrMsg != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMetricsAgent_compressData(t *testing.T) {
	tests := []struct {
		name           string
		input          []byte
		expectError    bool
		expectedErrMsg string
	}{
		{
			name:           "success",
			input:          []byte("test data for gzip compression"),
			expectError:    false,
			expectedErrMsg: "",
		},
		{
			name:           "empty data",
			input:          []byte{},
			expectError:    false,
			expectedErrMsg: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &MetricsAgent{}

			buf, err := a.compressData(tt.input)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, buf)
				if tt.expectedErrMsg != "" {
					assert.Contains(t, err.Error(), tt.expectedErrMsg)
				}
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, buf)

			r, err := gzip.NewReader(bytes.NewReader(buf.Bytes()))
			assert.NoError(t, err)
			defer r.Close()

			out, err := io.ReadAll(r)
			assert.NoError(t, err)
			assert.Equal(t, tt.input, out)
		})
	}
}

func TestMetricsAgent_prepareMetric(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*metric.MockMetric)
		expectedMetric *models.Metrics
		expectedErrMsg string
	}{
		{
			name: "valid counter metric",
			setupMock: func(m *metric.MockMetric) {
				m.EXPECT().Name().Return("requests_count")
				m.EXPECT().Type().Return(models.Counter)
				m.EXPECT().Value().Return(int64(42))
			},
			expectedMetric: &models.Metrics{
				ID:    "requests_count",
				MType: string(models.Counter),
				Delta: func() *int64 { v := int64(42); return &v }(),
			},
		},
		{
			name: "valid gauge metric",
			setupMock: func(m *metric.MockMetric) {
				m.EXPECT().Name().Return("cpu_usage")
				m.EXPECT().Type().Return(models.Gauge)
				m.EXPECT().Value().Return(0.99)
			},
			expectedMetric: &models.Metrics{
				ID:    "cpu_usage",
				MType: string(models.Gauge),
				Value: func() *float64 { v := 0.99; return &v }(),
			},
		},
		{
			name: "invalid counter value type",
			setupMock: func(m *metric.MockMetric) {
				m.EXPECT().Name().Return("bad_counter")
				m.EXPECT().Type().Return(models.Counter)
				m.EXPECT().Value().Return("not an int64")
			},
			expectedErrMsg: "invalid delta value",
		},
		{
			name: "invalid gauge value type",
			setupMock: func(m *metric.MockMetric) {
				m.EXPECT().Name().Return("bad_gauge")
				m.EXPECT().Type().Return(models.Gauge)
				m.EXPECT().Value().Return("not a float64")
			},
			expectedErrMsg: "invalid value",
		},
		{
			name: "unknown metric type",
			setupMock: func(m *metric.MockMetric) {
				m.EXPECT().Name().Return("unknown_metric")
				m.EXPECT().Type().Return("custom")
				m.EXPECT().Value().Return(123)
			},
			expectedErrMsg: "unknown metric type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMetric := metric.NewMockMetric(t)
			tt.setupMock(mockMetric)

			a := &MetricsAgent{}
			result, err := a.prepareMetric(mockMetric)

			if tt.expectedErrMsg != "" {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedMetric, result)
			}
		})
	}
}
