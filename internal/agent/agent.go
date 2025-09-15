package agent

import (
	"fmt"
	"io"
	"runtime"
	"sync"

	"github.com/gabkaclassic/metrics/pkg/httpclient"
	"github.com/gabkaclassic/metrics/pkg/metric"
	"log/slog"
)

type Agent interface {
	Poll()
	Report() error
}

type MetricsAgent struct {
	Agent
	stats   *runtime.MemStats
	mu      sync.RWMutex
	client  *httpclient.Client
	metrics []metric.Metric
}

func NewAgent(client *httpclient.Client, stats *runtime.MemStats) *MetricsAgent {
	metrics := []metric.Metric{
		// Counters
		&metric.PollCount{},

		// Gauges
		&metric.RandomValue{},
	}

	// Gauges runtime
	metrics = append(metrics, metric.RuntimeMetrics(stats)...)

	return &MetricsAgent{
		client:  client,
		metrics: metrics,
		stats:   stats,
	}
}

func (agent *MetricsAgent) Poll() {
	runtime.ReadMemStats(agent.stats)
	for _, metric := range agent.metrics {
		metric.Update()
	}
}

func (agent *MetricsAgent) Report() error {
	agent.mu.RLock()
	metricCopy := make([]metric.Metric, len(agent.metrics))
	copy(metricCopy, agent.metrics)
	agent.mu.RUnlock()

	var wg sync.WaitGroup
	errCh := make(chan error, len(metricCopy))
	resultCh := make(chan string, len(metricCopy))

	for _, currentMetric := range metricCopy {
		wg.Add(1)
		go func(m metric.Metric) {
			defer wg.Done()

			url := fmt.Sprintf("/%s/%s/%v",
				m.Type(), m.Name(), m.Value(),
			)

			respCh, errCh := agent.client.Post(url, nil)

			select {
			case resp := <-respCh:
				if resp != nil {
					defer resp.Body.Close()
					body, err := io.ReadAll(resp.Body)
					if err != nil {
						slog.Error(
							"Read sending report body error",
							slog.Any("metric", m),
							slog.Any("response", resp),
							slog.String("error", err.Error()),
						)
						return
					}
					resultCh <- fmt.Sprintf("Metric %s: %s", m.Name(), string(body))
				}
			case err := <-errCh:
				slog.Error(
					"Send metric report error",
					slog.Any("metric", m),
					slog.String("error", err.Error()),
				)
			}
		}(currentMetric)
	}

	go func() {
		wg.Wait()
		close(errCh)
		close(resultCh)
	}()

	for result := range resultCh {
		slog.Info(
			"Send metric report result",
			slog.Any("result", result),
		)
	}

	var errors []error
	for err := range errCh {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return fmt.Errorf("completed with %d errors: %v", len(errors), errors)
	}

	return nil
}
