package repository

import (
	"errors"
	"sync"

	"github.com/gabkaclassic/metrics/internal/model"
	"github.com/gabkaclassic/metrics/internal/storage"
	"github.com/gabkaclassic/metrics/pkg/metric"
)

type MetricsRepository interface {
	Add(metric models.Metrics) error
	Reset(metric models.Metrics) error
	Get(metricID string) (*models.Metrics, error)
	GetAll() *map[string]any
}

type metricsRepository struct {
	storage *storage.MemStorage
	mutex   *sync.RWMutex
}

func NewMetricsRepository(storage *storage.MemStorage) MetricsRepository {

	if storage == nil {
		panic(errors.New("create new metrics repository failed: storage is nil"))
	}

	return &metricsRepository{
		storage: storage,
		mutex:   &sync.RWMutex{},
	}
}

func (repository *metricsRepository) GetAll() *map[string]any {

	metrics := make(map[string]any, len(repository.storage.Metrics))

	for id, m := range repository.storage.Metrics {
		switch m.MType {
		case string(metric.CounterType):
			metrics[id] = *m.Delta
		case string(metric.GaugeType):
			metrics[id] = *m.Value
		}
	}

	return &metrics
}

func (repository *metricsRepository) Get(metricID string) (*models.Metrics, error) {

	metric, exists := repository.storage.Metrics[metricID]

	if !exists {
		return nil, nil
	}

	return &metric, nil
}

func (repository *metricsRepository) updateMetric(metric models.Metrics, updateMetricFunction func(metric models.Metrics) error) error {
	repository.mutex.Lock()
	defer repository.mutex.Unlock()

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
