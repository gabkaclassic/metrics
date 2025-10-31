package agent

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"sync"

	"encoding/json"
	"log/slog"

	"compress/gzip"
	"github.com/gabkaclassic/metrics/internal/model"
	"github.com/gabkaclassic/metrics/pkg/hash"
	"github.com/gabkaclassic/metrics/pkg/httpclient"
	"github.com/gabkaclassic/metrics/pkg/metric"
)

type Agent interface {
	Poll()
	Report() error
}

type MetricsAgent struct {
	Agent
	stats          *runtime.MemStats
	mu             *sync.RWMutex
	client         httpclient.HTTPClient
	metrics        []metric.Metric
	batchesEnabled bool
	signer         hash.Signer
}

func NewAgent(client httpclient.HTTPClient, stats *runtime.MemStats, batchesEnabled bool, signKey string) *MetricsAgent {
	metrics := []metric.Metric{
		// Counters
		&metric.PollCount{},

		// Gauges
		&metric.RandomValue{},
	}

	// Gauges runtime
	metrics = append(metrics, metric.RuntimeMetrics(stats)...)

	signer := hash.NewSHA256Signer(signKey)

	return &MetricsAgent{
		client:         client,
		metrics:        metrics,
		stats:          stats,
		mu:             &sync.RWMutex{},
		batchesEnabled: batchesEnabled,
		signer:         signer,
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

	if agent.batchesEnabled {
		return agent.reportBatch(metricCopy)
	}
	return agent.reportIndividual(metricCopy)
}

func (agent *MetricsAgent) reportIndividual(metrics []metric.Metric) error {
	var wg sync.WaitGroup
	errCh := make(chan error, len(metrics))

	for _, m := range metrics {
		wg.Add(1)
		go func(metricEntity metric.Metric) {
			defer wg.Done()

			metricModel, err := agent.prepareMetric(metricEntity)
			if err != nil {
				slog.Error("Prepare metric error", slog.Any("metric", metricEntity), slog.String("error", err.Error()))
				errCh <- err
				return
			}

			raw, err := json.Marshal(metricModel)
			if err != nil {
				slog.Error("Marshal metric error", slog.Any("metric", metricEntity), slog.String("error", err.Error()))
				errCh <- err
				return
			}

			buffer, err := agent.compressData(raw)
			if err != nil {
				slog.Error("Compress data error", slog.Any("metric", metricEntity), slog.String("error", err.Error()))
				errCh <- err
				return
			}

			if err := agent.sendRequest("/update/", buffer); err != nil {
				slog.Error("Send metric error", slog.Any("metric", metricEntity), slog.String("error", err.Error()))
				errCh <- err
			}

		}(m)
	}

	go func() {
		wg.Wait()
		close(errCh)
	}()

	var errors []error
	for err := range errCh {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return fmt.Errorf("completed with %d errors: %v", len(errors), errors)
	}

	slog.Info("All individual metrics sent successfully", slog.Int("count", len(metrics)))
	return nil
}

func (agent *MetricsAgent) reportBatch(metrics []metric.Metric) error {
	var metricModels []models.Metrics

	for _, m := range metrics {
		metricModel, err := agent.prepareMetric(m)
		if err != nil {
			return fmt.Errorf("prepare metric %s error: %w", m.Name(), err)
		}
		metricModels = append(metricModels, *metricModel)
	}

	raw, err := json.Marshal(metricModels)
	if err != nil {
		return fmt.Errorf("marshal metrics batch error: %w", err)
	}

	buffer, err := agent.compressData(raw)
	if err != nil {
		return fmt.Errorf("compress batch data error: %w", err)
	}

	if err := agent.sendRequest("/updates/", buffer); err != nil {
		return fmt.Errorf("send metrics batch error: %w", err)
	}

	slog.Info("Metrics batch sent successfully", slog.Int("count", len(metrics)))
	return nil
}

func (agent *MetricsAgent) prepareMetric(m metric.Metric) (*models.Metrics, error) {
	metricName := m.Name()
	metricType := m.Type()
	metricRawValue := m.Value()

	metricModel := &models.Metrics{
		ID:    metricName,
		MType: string(metricType),
	}

	switch metricType {
	case models.Counter:
		delta, ok := metricRawValue.(int64)
		if !ok {
			return nil, fmt.Errorf("invalid delta value: %v", metricRawValue)
		}
		metricModel.Delta = &delta
	case models.Gauge:
		value, ok := metricRawValue.(float64)
		if !ok {
			return nil, fmt.Errorf("invalid value: %v", metricRawValue)
		}
		metricModel.Value = &value
	default:
		return nil, fmt.Errorf("unknown metric type: %s", metricType)
	}

	return metricModel, nil
}

func (agent *MetricsAgent) compressData(data []byte) (*bytes.Buffer, error) {
	var buffer bytes.Buffer
	compressor := gzip.NewWriter(&buffer)

	if _, err := compressor.Write(data); err != nil {
		compressor.Close()
		return nil, fmt.Errorf("compressing data error: %w", err)
	}

	if err := compressor.Close(); err != nil {
		return nil, fmt.Errorf("closing compressor failed: %w", err)
	}

	return &buffer, nil
}

func (agent *MetricsAgent) sendRequest(endpoint string, body *bytes.Buffer) error {

	sign := agent.signer.Sign(body.Bytes())

	resp, err := agent.client.Post(
		endpoint,
		&httpclient.RequestOptions{
			Body: body,
			Headers: &httpclient.Headers{
				"Content-Type":     "application/json",
				"Content-Encoding": "gzip",
				"Hash":             sign,
			},
		},
	)

	if err != nil {
		return fmt.Errorf("send request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		responseBody, err := io.ReadAll(resp.Body)

		if err != nil {
			return err
		}

		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(responseBody))
	}

	responseBody, err := io.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	slog.Info("Request completed successfully",
		slog.String("endpoint", endpoint),
		slog.String("response", string(responseBody)),
	)

	return nil
}
