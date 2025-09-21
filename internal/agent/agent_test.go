package agent

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"runtime"
	"testing"

	"github.com/gabkaclassic/metrics/pkg/httpclient"
	"github.com/gabkaclassic/metrics/pkg/metric"
	"github.com/stretchr/testify/assert"
)

func TestNewAgent(t *testing.T) {
	dummyClient := &httpclient.Client{}
	stats := &runtime.MemStats{}

	agent := NewAgent(dummyClient, stats)

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

type mockMetric struct {
	updated bool
}

func (m *mockMetric) Update()                 { m.updated = true }
func (m *mockMetric) Name() string            { return "mock" }
func (m *mockMetric) Value() any              { return "mock" }
func (m *mockMetric) Type() metric.MetricType { return metric.GaugeType }

func TestMetricsAgent_Poll(t *testing.T) {
	stats := &runtime.MemStats{}
	agent := &MetricsAgent{
		stats: stats,
		metrics: []metric.Metric{
			&mockMetric{},
			&mockMetric{},
		},
	}

	agent.Poll()

	assert.NotZero(t, agent.stats.Alloc)

	for _, m := range agent.metrics {
		mock := m.(*mockMetric)
		assert.True(t, mock.updated)
	}
}

type mockHTTPClient struct {
	postFunc func(url string, opts *httpclient.RequestOptions) (*http.Response, error)
}

func (m *mockHTTPClient) Get(url string, opts *httpclient.RequestOptions) (*http.Response, error) {
	panic("not implemented")
}
func (m *mockHTTPClient) Post(url string, opts *httpclient.RequestOptions) (*http.Response, error) {
	return m.postFunc(url, opts)
}
func (m *mockHTTPClient) Put(url string, opts *httpclient.RequestOptions) (*http.Response, error) {
	panic("not implemented")
}
func (m *mockHTTPClient) Patch(url string, opts *httpclient.RequestOptions) (*http.Response, error) {
	panic("not implemented")
}
func (m *mockHTTPClient) Delete(url string, opts *httpclient.RequestOptions) (*http.Response, error) {
	panic("not implemented")
}

func TestMetricsAgent_Report(t *testing.T) {
	tests := []struct {
		name       string
		clientFunc func() httpclient.HttpClient
		wantErr    bool
	}{
		{
			name: "success",
			clientFunc: func() httpclient.HttpClient {
				return &mockHTTPClient{
					postFunc: func(url string, opts *httpclient.RequestOptions) (*http.Response, error) {
						return &http.Response{
							StatusCode: 200,
							Body:       io.NopCloser(bytes.NewBufferString("ok")),
						}, nil
					},
				}
			},
			wantErr: false,
		},
		{
			name: "error",
			clientFunc: func() httpclient.HttpClient {
				return &mockHTTPClient{
					postFunc: func(url string, opts *httpclient.RequestOptions) (*http.Response, error) {
						return nil, errors.New("network error")
					},
				}
			},
			wantErr: true,
		},
		{
			name: "nil response",
			clientFunc: func() httpclient.HttpClient {
				return &mockHTTPClient{
					postFunc: func(url string, opts *httpclient.RequestOptions) (*http.Response, error) {
						return nil, nil
					},
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := &MetricsAgent{
				client: tt.clientFunc(),
				stats:  &runtime.MemStats{},
				metrics: []metric.Metric{
					&mockMetric{},
				},
			}

			err := agent.Report()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
