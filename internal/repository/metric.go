package repository

import (
	"errors"

	"github.com/gabkaclassic/metrics/internal/model"
	"github.com/gabkaclassic/metrics/internal/storage"
)

type MetricsRepository interface {
	Add(metric models.Metrics) error
	Reset(metric models.Metrics) error
	Get(metricID string) (*models.Metrics, error)
}

type metricsRepository struct {
	MetricsRepository

	storage *storage.MemStorage
}

func NewMetricsRepository(storage *storage.MemStorage) MetricsRepository {

	if storage == nil {
		panic(errors.New("create new metrics repository failed: storage is nil"))
	}

	return &metricsRepository{
		storage: storage,
	}
}

func (repository *metricsRepository) Get(metricID string) (*models.Metrics, error) {

	metric, exists := repository.storage.Metrics[metricID]

	if !exists {
		return nil, nil
	}

	return &metric, nil
}

func (repository *metricsRepository) updateMetric(metric models.Metrics, updateMetricFunction func(metric models.Metrics) error) error {
	repository.storage.Mutex.Lock()
	defer repository.storage.Mutex.Unlock()

	if repository.storage.Metrics == nil {
		repository.storage.Metrics = make(map[string]models.Metrics)
	}

	err := updateMetricFunction(metric)

	return err
}

func (repository *metricsRepository) Add(metric models.Metrics) error {

	err := repository.updateMetric(
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

func (repository *metricsRepository) Reset(metric models.Metrics) error {

	err := repository.updateMetric(
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
