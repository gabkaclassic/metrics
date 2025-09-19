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

type mockHttpClient struct {
	postFunc func(url string, body io.Reader) (<-chan *http.Response, <-chan error)
}

func (m *mockHttpClient) Get(url string, params httpclient.Params) (<-chan *http.Response, <-chan error) {
	panic("not implemented")
}
func (m *mockHttpClient) Post(url string, body io.Reader) (<-chan *http.Response, <-chan error) {
	return m.postFunc(url, body)
}
func (m *mockHttpClient) Put(url string, body io.Reader) (<-chan *http.Response, <-chan error) {
	panic("not implemented")
}
func (m *mockHttpClient) Patch(url string, body io.Reader) (<-chan *http.Response, <-chan error) {
	panic("not implemented")
}
func (m *mockHttpClient) Delete(url string, params httpclient.Params, body io.Reader) (<-chan *http.Response, <-chan error) {
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
				return &mockHttpClient{
					postFunc: func(url string, body io.Reader) (<-chan *http.Response, <-chan error) {
						respCh := make(chan *http.Response, 1)
						errCh := make(chan error, 1)
						respCh <- &http.Response{
							StatusCode: 200,
							Body:       io.NopCloser(bytes.NewBufferString("ok")),
						}
						return respCh, errCh
					},
				}
			},
			wantErr: false,
		},
		{
			name: "error",
			clientFunc: func() httpclient.HttpClient {
				return &mockHttpClient{
					postFunc: func(url string, body io.Reader) (<-chan *http.Response, <-chan error) {
						errCh := make(chan error, 1)
						errCh <- errors.New("network error")
						return nil, errCh
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
