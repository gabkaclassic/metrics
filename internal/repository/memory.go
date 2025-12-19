package repository

import (
	"context"
	"errors"
	"fmt"
	"sync"

	models "github.com/gabkaclassic/metrics/internal/model"
	"github.com/gabkaclassic/metrics/internal/storage"
	"github.com/gabkaclassic/metrics/pkg/metric"
)

// MetricsRepository defines the interface for metric data operations.
// Implementations provide persistence-agnostic access to metrics.
type MetricsRepository interface {
	// Add increments a counter metric or adds a new metric.
	// For counter metrics, adds delta to existing value.
	// For gauge metrics, adds new metric if not exists.
	Add(context.Context, models.Metrics) error

	// AddAll performs batch addition of metrics.
	// More efficient than multiple Add calls for bulk operations.
	AddAll(context.Context, []models.Metrics) error

	// ResetAll performs batch reset of gauge metrics.
	// Updates existing gauge values or adds new ones.
	ResetAll(context.Context, []models.Metrics) error

	// Reset sets a gauge metric to a specific value.
	// Creates the metric if it doesn't exist.
	Reset(context.Context, models.Metrics) error

	// Get retrieves a single metric by its ID.
	// Returns error if metric not found.
	Get(context.Context, string) (*models.Metrics, error)

	// GetAll returns all metrics as a map of ID to value.
	// Counter metrics return int64, gauge metrics return float64.
	GetAll(context.Context) (map[string]any, error)

	// GetAllMetrics returns all metrics as a slice of models.Metrics.
	// Preserves complete metric structure including type and hash.
	GetAllMetrics(context.Context) ([]models.Metrics, error)
}

// memoryMetricsRepository implements MetricsRepository using in-memory storage.
// Provides thread-safe operations through read-write mutex.
// Suitable for single-instance deployments and testing.
type memoryMetricsRepository struct {
	storage *storage.MemStorage
	mutex   *sync.RWMutex
}

// NewMemoryMetricsRepository creates a new in-memory metrics repository.
//
// storage: MemStorage instance for data persistence
// mutex: Read-write mutex for thread safety (can be shared)
//
// Returns:
//   - MetricsRepository: Ready-to-use repository instance
//   - error: If storage is nil
//
// Note: The repository uses the provided mutex for synchronization.
// For independent synchronization, create new sync.RWMutex.
func NewMemoryMetricsRepository(storage *storage.MemStorage, mutex *sync.RWMutex) (MetricsRepository, error) {
	if storage == nil {
		return nil, errors.New("create new metrics repository failed: storage is nil")
	}

	return &memoryMetricsRepository{
		storage: storage,
		mutex:   mutex,
	}, nil
}

// GetAllMetrics returns all stored metrics as a slice.
// Order of metrics in the slice is not guaranteed.
func (repository *memoryMetricsRepository) GetAllMetrics(ctx context.Context) ([]models.Metrics, error) {
	metrics := make([]models.Metrics, len(repository.storage.Metrics))
	index := 0
	for _, m := range repository.storage.Metrics {
		metrics[index] = m
		index += 1
	}

	return metrics, nil
}

// GetAll returns all metrics as a map of metric ID to value.
// Counter metrics are returned as int64, gauge metrics as float64.
func (repository *memoryMetricsRepository) GetAll(ctx context.Context) (map[string]any, error) {
	metrics := make(map[string]any, len(repository.storage.Metrics))

	for id, m := range repository.storage.Metrics {
		switch m.MType {
		case string(metric.CounterType):
			metrics[id] = *m.Delta
		case string(metric.GaugeType):
			metrics[id] = *m.Value
		}
	}

	return metrics, nil
}

// Get retrieves a metric by its ID.
// Returns error if metric with given ID doesn't exist.
func (repository *memoryMetricsRepository) Get(ctx context.Context, metricID string) (*models.Metrics, error) {
	metric, exists := repository.storage.Metrics[metricID]

	if !exists {
		return nil, fmt.Errorf("metric %s not found", metricID)
	}

	return &metric, nil
}

// updateMetric executes a metric update operation with thread safety.
// Acquires write lock and ensures storage map is initialized.
func (repository *memoryMetricsRepository) updateMetric(_ context.Context, metric models.Metrics, updateMetricFunction func(metric models.Metrics) error) error {
	repository.mutex.Lock()
	defer repository.mutex.Unlock()

	if repository.storage.Metrics == nil {
		repository.storage.Metrics = make(map[string]models.Metrics)
	}

	err := updateMetricFunction(metric)

	return err
}

// updateMetrics executes a batch metrics update with thread safety.
// Acquires write lock and ensures storage map is initialized.
func (repository *memoryMetricsRepository) updateMetrics(_ context.Context, metrics []models.Metrics, updateMetricsFunction func(metric []models.Metrics) error) error {
	repository.mutex.Lock()
	defer repository.mutex.Unlock()

	if repository.storage.Metrics == nil {
		repository.storage.Metrics = make(map[string]models.Metrics)
	}

	err := updateMetricsFunction(metrics)

	return err
}

// Add increments a counter metric or adds a new metric.
// For counters: adds delta to existing value (creates if not exists)
// For gauges: adds new metric if not exists (no increment)
func (repository *memoryMetricsRepository) Add(ctx context.Context, metric models.Metrics) error {
	err := repository.updateMetric(
		ctx,
		metric,
		func(metric models.Metrics) error {
			if savedMetric, exists := repository.storage.Metrics[metric.ID]; exists {
				*savedMetric.Delta = *(savedMetric.Delta) + *(metric.Delta)
			} else {
				repository.storage.Metrics[metric.ID] = metric
			}
			return nil
		},
	)

	return err
}

// AddAll performs batch addition of metrics.
// More efficient than individual Add calls for multiple metrics.
func (repository *memoryMetricsRepository) AddAll(ctx context.Context, metrics []models.Metrics) error {
	err := repository.updateMetrics(
		ctx,
		metrics,
		func(metrics []models.Metrics) error {
			for _, metric := range metrics {
				if savedMetric, exists := repository.storage.Metrics[metric.ID]; exists {
					*savedMetric.Delta = *(savedMetric.Delta) + *(metric.Delta)
				} else {
					repository.storage.Metrics[metric.ID] = metric
				}
			}
			return nil
		},
	)

	return err
}

// Reset sets a gauge metric to a specific value.
// Creates the metric if it doesn't exist.
func (repository *memoryMetricsRepository) Reset(ctx context.Context, metric models.Metrics) error {
	err := repository.updateMetric(
		ctx,
		metric,
		func(metric models.Metrics) error {
			if savedMetric, exists := repository.storage.Metrics[metric.ID]; exists {
				*savedMetric.Value = *(metric.Value)
			} else {
				repository.storage.Metrics[metric.ID] = metric
			}
			return nil
		},
	)

	return err
}

// ResetAll performs batch reset of gauge metrics.
// Updates existing values or adds new metrics.
func (repository *memoryMetricsRepository) ResetAll(ctx context.Context, metrics []models.Metrics) error {
	err := repository.updateMetrics(
		ctx,
		metrics,
		func(metrics []models.Metrics) error {
			for _, metric := range metrics {
				if savedMetric, exists := repository.storage.Metrics[metric.ID]; exists {
					*savedMetric.Value = *(metric.Value)
				} else {
					repository.storage.Metrics[metric.ID] = metric
				}
			}
			return nil
		},
	)

	return err
}
