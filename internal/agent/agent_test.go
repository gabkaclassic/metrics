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
	"github.com/stretchr/testify/mock"
)

func TestNewAgent(t *testing.T) {
	dummyClient := httpclient.NewMockHTTPClient(t)
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
func TestMetricsAgent_Report(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(m *httpclient.MockHTTPClient)
		wantErr bool
	}{
		{
			name: "success",
			setup: func(m *httpclient.MockHTTPClient) {
				m.EXPECT().
					Post(mock.Anything, mock.Anything).
					Return(&http.Response{
						StatusCode: 200,
						Body:       io.NopCloser(bytes.NewBufferString("ok")),
					}, nil)
			},
			wantErr: false,
		},
		{
			name: "error",
			setup: func(m *httpclient.MockHTTPClient) {
				m.EXPECT().
					Post(mock.Anything, mock.Anything).
					Return(nil, errors.New("network error"))
			},
			wantErr: true,
		},
		{
			name: "nil response",
			setup: func(m *httpclient.MockHTTPClient) {
				m.EXPECT().
					Post(mock.Anything, mock.Anything).
					Return(nil, nil)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := httpclient.NewMockHTTPClient(t)

			tt.setup(mockClient)

			agent := NewAgent(mockClient, &runtime.MemStats{})

			err := agent.Report()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
