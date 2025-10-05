package agent

import (
	"bytes"
	"fmt"
	"io"
	"runtime"
	"sync"

	"encoding/json"
	"log/slog"

	"github.com/gabkaclassic/metrics/internal/model"
	"github.com/gabkaclassic/metrics/pkg/httpclient"
	"github.com/gabkaclassic/metrics/pkg/metric"
)

type Agent interface {
	Poll()
	Report() error
}

type MetricsAgent struct {
	Agent
	stats   *runtime.MemStats
	mu      *sync.RWMutex
	client  httpclient.HTTPClient
	metrics []metric.Metric
}

func NewAgent(client httpclient.HTTPClient, stats *runtime.MemStats) *MetricsAgent {
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
		mu:      &sync.RWMutex{},
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

	for _, m := range metricCopy {
		wg.Add(1)
		go func(metricEntity metric.Metric) {
			defer wg.Done()

			metricName := metricEntity.Name()
			metricType := metricEntity.Type()
			metricRawValue := metricEntity.Value()

			metricModel := models.Metrics{
				ID:    metricName,
				MType: string(metricType),
			}

			switch metricType {
			case models.Counter:
				delta, ok := metricRawValue.(int64)

				if !ok {
					errCh <- fmt.Errorf("invalid delta value: %v", metricRawValue)
					return
				}
				metricModel.Delta = &delta
			case models.Gauge:
				value, ok := metricRawValue.(float64)

				if !ok {
					errCh <- fmt.Errorf("invalid value: %v", metricRawValue)
					return
				}
				metricModel.Value = &value
			}

			raw, err := json.Marshal(metricModel)

			if err != nil {
				slog.Error(
					"Marshal metric report error",
					slog.Any("metric", metricEntity),
					slog.String("error", err.Error()),
				)
				errCh <- err
				return
			}

			resp, err := agent.client.Post(
				"/update/",
				&httpclient.RequestOptions{
					Body: bytes.NewBuffer(raw),
					Headers: &httpclient.Headers{
						"Content-Type": "application/json",
					},
				},
			)

			if err != nil {
				slog.Error(
					"Send metric report error",
					slog.Any("metric", metricEntity),
					slog.String("error", err.Error()),
				)
				errCh <- err
				return
			}

			if resp == nil {
				errCh <- fmt.Errorf("nil response for metric %s", metricName)
				return
			}

			defer resp.Body.Close()
			body, readErr := io.ReadAll(resp.Body)
			if readErr != nil {
				slog.Error(
					"Read sending report body error",
					slog.Any("metric", metricEntity),
					slog.Any("response", resp),
					slog.String("error", readErr.Error()),
				)
				errCh <- readErr
				return
			}

			resultCh <- fmt.Sprintf("Metric %s: %s", metricName, string(body))
		}(m)
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
