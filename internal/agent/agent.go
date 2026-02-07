// Package agent implements metrics collection and reporting agent.
//
// The agent periodically collects system metrics (runtime, CPU, memory, custom)
// and reports them to a metrics server. It supports multiple reporting strategies:
//   - Individual metric reporting (one HTTP request per metric)
//   - Batch reporting (multiple metrics per HTTP request)
//   - Rate-limited concurrent reporting
//
// Metrics collected include:
//   - Go runtime statistics (memory allocation, GC, etc.)
//   - System metrics (CPU utilization, memory usage)
//   - Custom application metrics (poll count, random values)
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
	"github.com/gabkaclassic/metrics/pkg/crypt"
	"github.com/gabkaclassic/metrics/pkg/hash"
	"github.com/gabkaclassic/metrics/pkg/httpclient"
	"github.com/gabkaclassic/metrics/pkg/metric"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

// Agent defines the interface for metrics collection agents.
// Implementations handle periodic polling and reporting of metrics.
type Agent interface {
	// Poll collects current metric values from all sources.
	// Updates internal metric states with fresh values.
	Poll()

	// Report sends collected metrics to the server.
	// Returns error if any part of the reporting process fails.
	Report() error
}

// MetricsAgent implements the Agent interface with advanced features.
// Provides concurrent collection, batch reporting, and rate limiting.
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
	encryptor      *crypt.Encryptor
}

// NewAgent creates and initializes a new metrics collection agent.
//
// client: HTTP client configured with server endpoint.
// batchesEnabled: Enables batch reporting when true.
// signKey: Secret key for request signature generation.
// rateLimit: Maximum concurrent HTTP requests (0 for no limit).
// batchSize: Maximum metrics per batch (ignored if batchesEnabled false).
//
// Returns:
//   - *MetricsAgent: Fully initialized agent ready for polling
//   - error: If system metric collection fails during initialization
//
// The agent initializes with:
//   - Default metrics (PollCount, RandomValue)
//   - Go runtime metrics
//   - System metrics (CPU, memory)
//   - Request signer for secure communication
//   - Request encryptor for requests
func NewAgent(client httpclient.HTTPClient, batchesEnabled bool, signKey string, publicKeyPath string, rateLimit int, batchSize int) (*MetricsAgent, error) {
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

	if len(publicKeyPath) > 0 {
		agent.encryptor, err = crypt.NewEncryptor(publicKeyPath)

		if err != nil {
			return nil, err
		}
	} else {
		agent.encryptor = nil
	}

	return agent, nil
}

// Poll collects current values for all metrics.
// Updates internal metric states and queues them for reporting.
// Non-blocking: if previous report is in progress, new metrics are skipped.
// Collected metrics include:
//   - Go runtime memory statistics (via runtime.ReadMemStats)
//   - CPU utilization (via gopsutil)
//   - System memory (via gopsutil)
//   - All registered custom metrics
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

// Report sends queued metrics to the server.
// Uses either batch or individual reporting based on configuration.
// Non-blocking: returns immediately if no metrics are queued.
// Returns error if reporting fails for any metric.
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

// reportIndividual sends metrics one-by-one with concurrent workers.
// Uses worker pool pattern with configurable rate limit.
// metrics: Slice of metrics to send individually.
// Returns combined error if any worker encounters errors.
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

// reportWorker is a goroutine that processes individual metric reporting jobs.
// wg: WaitGroup for coordinating worker shutdown.
// jobs: Channel receiving metrics to report.
// errCh: Channel for reporting errors (non-blocking).
// Each worker handles metrics sequentially until jobs channel closes.
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

// reportBatchWithLimit sends metrics in batches with rate limiting.
// Splits metrics into chunks and processes them concurrently up to rate limit.
// metrics: All metrics to send in batches.
// Returns combined error if any batch fails.
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

// chunkMetrics splits a slice of metrics into chunks of specified size.
// metrics: Slice to split.
// size: Maximum chunk size.
// Returns 2D slice where each inner slice has length <= size.
func chunkMetrics(metrics []metric.Metric, size int) [][]metric.Metric {
	var chunks [][]metric.Metric
	for size < len(metrics) {
		metrics, chunks = metrics[size:], append(chunks, metrics[0:size:size])
	}
	chunks = append(chunks, metrics)
	return chunks
}

// reportBatch sends a single batch of metrics in one HTTP request.
// metrics: Metrics to include in this batch.
// Returns error if any step fails (preparation, marshaling, sending).
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

// prepareMetric converts a metric.Metric to models.Metrics for transmission.
// m: Source metric with current value.
// Returns JSON-serializable model with appropriate value fields set.
// Returns error for unknown metric types or value conversion failures.
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

// compressData compresses JSON data using gzip for network transmission.
// data: Raw JSON bytes to compress.
// Returns buffer containing gzipped data.
// Returns error if compression fails.
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

func (agent *MetricsAgent) prepareRequestBody(body *bytes.Buffer) (*bytes.Buffer, string, error) {
	data := body.Bytes()

	if agent.encryptor != nil {
		encrypted, err := agent.encryptor.Encrypt(data)
		if err != nil {
			return nil, "", fmt.Errorf("encrypt error: %w", err)
		}
		body.Reset()
		body.Write(encrypted)
		data = encrypted
	}

	sign := agent.signer.Sign(data)

	return body, sign, nil
}

// sendRequest sends an HTTP POST request with compressed, signed data.
// endpoint: Server endpoint path (e.g., "/update/" or "/updates/").
// body: Compressed and encrypt request body.
// Returns error if request fails or server returns non-200 status.
// Automatically adds required headers: Content-Type, Content-Encoding, Hash.
func (agent *MetricsAgent) sendRequest(endpoint string, body *bytes.Buffer) error {

	body, sign, err := agent.prepareRequestBody(body)

	if err != nil {
		return fmt.Errorf("prepare request body error: %w", err)
	}

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
	slog.Info("Request completed successfully", slog.String("endpoint", endpoint), slog.String("response", string(responseBody)))
	return nil
}

// dispatchReports routes metrics to appropriate reporting method.
// Internal method used by StartReporting for consistent dispatch logic.
// metrics: Collected metrics to report.
// Returns error if reporting fails.
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

// StartReporting runs the continuous reporting loop.
// Listens for metrics on jobCh and dispatches them for reporting.
// ctx: Context for graceful shutdown.
// Runs until context is cancelled, ensuring in-progress reports complete.
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
