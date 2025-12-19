package agent

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"sync"
	"time"

	"encoding/json"
	"log/slog"

	"compress/gzip"

	models "github.com/gabkaclassic/metrics/internal/model"
	"github.com/gabkaclassic/metrics/pkg/hash"
	"github.com/gabkaclassic/metrics/pkg/httpclient"
	"github.com/gabkaclassic/metrics/pkg/metric"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

type Agent interface {
	Poll()
	Report() error
}

type MetricsAgent struct {
	Agent
	stats          *runtime.MemStats
	psMemStats     *mem.VirtualMemoryStat
	cpuStats       *[]float64
	mu             *sync.RWMutex
	client         httpclient.HTTPClient
	metrics        []metric.Metric
	batchesEnabled bool
	signer         hash.Signer
	rateLimit      int
	jobCh          chan []metric.Metric
	wg             sync.WaitGroup
	batchSize      int
}

func NewAgent(client httpclient.HTTPClient, batchesEnabled bool, signKey string, rateLimit int, batchSize int) (*MetricsAgent, error) {
	metrics := []metric.Metric{
		// Counters
		&metric.PollCount{},

		// Gauges
		&metric.RandomValue{},
	}

	// Gauges runtime
	stats := &runtime.MemStats{}
	metrics = append(metrics, metric.RuntimeMetrics(stats)...)

	agent := &MetricsAgent{
		client:         client,
		stats:          stats,
		mu:             &sync.RWMutex{},
		batchesEnabled: batchesEnabled,
		rateLimit:      rateLimit,
		jobCh:          make(chan []metric.Metric, 1),
		batchSize:      batchSize,
	}
	cpuStats, err := cpu.Percent(1*time.Second, false)

	if err != nil {
		return nil, err
	}

	psMemStats, err := mem.VirtualMemory()

	if err != nil {
		return nil, err
	}

	agent.cpuStats = &cpuStats
	agent.psMemStats = psMemStats

	metrics = append(metrics, metric.PsMetrics(
		func() *mem.VirtualMemoryStat { agent.mu.RLock(); defer agent.mu.RUnlock(); return agent.psMemStats },
		func() *[]float64 { agent.mu.RLock(); defer agent.mu.RUnlock(); return agent.cpuStats },
	)...)

	agent.metrics = metrics

	signer := hash.NewSHA256Signer(signKey)
	agent.signer = signer

	return agent, nil
}

func (agent *MetricsAgent) Poll() {
	slog.Debug("Start metrics polling")
	runtime.ReadMemStats(agent.stats)

	cpuStats, err := cpu.Percent(1*time.Second, false)
	if err != nil {
		slog.Error("Poll CPU metrics error", slog.Any("error", err))
	} else {
		agent.mu.Lock()
		agent.cpuStats = &cpuStats
		agent.mu.Unlock()
	}

	for _, metric := range agent.metrics {
		metric.Update()
	}

	agent.mu.RLock()
	metricCopy := make([]metric.Metric, len(agent.metrics))
	copy(metricCopy, agent.metrics)
	agent.mu.RUnlock()

	select {
	case agent.jobCh <- metricCopy:
		slog.Debug("Queued metrics for reporting")
	default:
		slog.Debug("Previous report still in progress, skipping new metrics batch")
	}
	slog.Debug("Finish metrics polling")
}

func (agent *MetricsAgent) Report() error {
	slog.Debug("Start metrics report")
	select {
	case metrics := <-agent.jobCh:
		if agent.batchesEnabled {
			return agent.reportBatchWithLimit(metrics)
		}
		return agent.reportIndividual(metrics)
	default:
		slog.Debug("No metrics to report")
		return nil
	}
}

func (agent *MetricsAgent) reportIndividual(metrics []metric.Metric) error {

	jobs := make(chan metric.Metric)
	errCh := make(chan error, len(metrics))

	var wg sync.WaitGroup

	wg.Add(agent.rateLimit)
	for i := 0; i < agent.rateLimit; i++ {
		go agent.reportWorker(&wg, jobs, errCh)
	}

	go func() {
		defer close(jobs)
		for _, m := range metrics {
			jobs <- m
		}
	}()

	wg.Wait()
	close(errCh)

	var errs []error
	for e := range errCh {
		errs = append(errs, e)
	}

	if len(errs) > 0 {
		return fmt.Errorf("completed with %d errors: %v", len(errs), errs)
	}

	slog.Info("All individual metrics sent successfully", slog.Int("count", len(metrics)))
	return nil
}

func (agent *MetricsAgent) reportWorker(wg *sync.WaitGroup, jobs <-chan metric.Metric, errCh chan<- error) {
	defer wg.Done()
	slog.Debug("Worker start")
	defer slog.Debug("Worker stop")

	for m := range jobs {
		metricModel, err := agent.prepareMetric(m)
		if err != nil {
			slog.Error("Prepare metric error", slog.Any("metric", m), slog.String("error", err.Error()))
			select {
			case errCh <- err:
			default:
			}
			return
		}

		raw, err := json.Marshal(metricModel)
		if err != nil {
			slog.Error("Marshal metric error", slog.Any("metric", m), slog.String("error", err.Error()))
			select {
			case errCh <- err:
			default:
			}
			return
		}

		buffer, err := agent.compressData(raw)
		if err != nil {
			slog.Error("Compress data error", slog.Any("metric", m), slog.String("error", err.Error()))
			select {
			case errCh <- err:
			default:
			}
			return
		}

		if err := agent.sendRequest("/update/", buffer); err != nil {
			slog.Error("Send metric error", slog.Any("metric", m), slog.String("error", err.Error()))
			select {
			case errCh <- err:
			default:
			}
			return
		}
	}
}

func (agent *MetricsAgent) reportBatchWithLimit(metrics []metric.Metric) error {

	chunks := chunkMetrics(metrics, agent.batchSize)

	sem := make(chan struct{}, agent.rateLimit)
	errCh := make(chan error, len(chunks))
	var wg sync.WaitGroup

	for _, chunk := range chunks {
		wg.Add(1)
		sem <- struct{}{}
		go func(chunk []metric.Metric) {
			defer wg.Done()
			defer func() { <-sem }()

			if err := agent.reportBatch(chunk); err != nil {
				errCh <- err
			}
		}(chunk)
	}

	wg.Wait()
	close(errCh)

	var errs []error
	for e := range errCh {
		errs = append(errs, e)
	}
	if len(errs) > 0 {
		return fmt.Errorf("batch sending completed with %d errors: %v", len(errs), errs)
	}

	slog.Info("All metric batches sent successfully", slog.Int("total", len(metrics)))
	return nil
}

func chunkMetrics(metrics []metric.Metric, size int) [][]metric.Metric {
	var chunks [][]metric.Metric
	for size < len(metrics) {
		metrics, chunks = metrics[size:], append(chunks, metrics[0:size:size])
	}
	chunks = append(chunks, metrics)
	return chunks
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

func (agent *MetricsAgent) dispatchReports(metrics []metric.Metric) error {
	if agent.batchesEnabled {
		return agent.reportBatch(metrics)
	}

	jobs := make(chan metric.Metric)
	errCh := make(chan error, len(metrics))

	for i := 0; i < agent.rateLimit; i++ {
		agent.wg.Add(1)
		go agent.reportWorker(&agent.wg, jobs, errCh)
	}

	go func() {
		defer close(jobs)
		for _, m := range metrics {
			jobs <- m
		}
	}()

	agent.wg.Wait()
	close(errCh)

	var errs []error
	for e := range errCh {
		errs = append(errs, e)
	}

	if len(errs) > 0 {
		return fmt.Errorf("completed with %d errors: %v", len(errs), errs)
	}
	return nil
}

func (agent *MetricsAgent) StartReporting(ctx context.Context) {
	for {
		select {
		case metrics := <-agent.jobCh:
			slog.Debug("Starting metrics report...")
			if err := agent.dispatchReports(metrics); err != nil {
				slog.Error("Report error", slog.String("error", err.Error()))
			} else {
				slog.Info("Report completed successfully")
			}
		case <-ctx.Done():
			slog.Info("Stopping reporter")
			return
		}
	}
}
