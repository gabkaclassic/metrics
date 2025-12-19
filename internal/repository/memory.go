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

type MetricsRepository interface {
	Add(context.Context, models.Metrics) error
	AddAll(context.Context, []models.Metrics) error
	ResetAll(context.Context, []models.Metrics) error
	Reset(context.Context, models.Metrics) error
	Get(context.Context, string) (*models.Metrics, error)
	GetAll(context.Context) (map[string]any, error)
	GetAllMetrics(context.Context) ([]models.Metrics, error)
}

type memoryMetricsRepository struct {
	storage *storage.MemStorage
	mutex   *sync.RWMutex
}

func NewMemoryMetricsRepository(storage *storage.MemStorage, mutex *sync.RWMutex) (MetricsRepository, error) {

	if storage == nil {
		return nil, errors.New("create new metrics repository failed: storage is nil")
	}

	return &memoryMetricsRepository{
		storage: storage,
		mutex:   mutex,
	}, nil
}

func (repository *memoryMetricsRepository) GetAllMetrics(ctx context.Context) ([]models.Metrics, error) {
	metrics := make([]models.Metrics, len(repository.storage.Metrics))
	index := 0
	for _, m := range repository.storage.Metrics {
		metrics[index] = m
		index += 1
	}

	return metrics, nil
}

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

func (repository *memoryMetricsRepository) Get(ctx context.Context, metricID string) (*models.Metrics, error) {

	metric, exists := repository.storage.Metrics[metricID]

	if !exists {
		return nil, fmt.Errorf("metric %s not found", metricID)
	}

	return &metric, nil
}

func (repository *memoryMetricsRepository) updateMetric(ctx context.Context, metric models.Metrics, updateMetricFunction func(metric models.Metrics) error) error {
	repository.mutex.Lock()
	defer repository.mutex.Unlock()

	if repository.storage.Metrics == nil {
		repository.storage.Metrics = make(map[string]models.Metrics)
	}

	err := updateMetricFunction(metric)

	return err
}
func (repository *memoryMetricsRepository) updateMetrics(ctx context.Context, metrics []models.Metrics, updateMetricsFunction func(metric []models.Metrics) error) error {
	repository.mutex.Lock()
	defer repository.mutex.Unlock()

	if repository.storage.Metrics == nil {
		repository.storage.Metrics = make(map[string]models.Metrics)
	}

	err := updateMetricsFunction(metrics)

	return err
}

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
