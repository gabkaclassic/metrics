package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"sync"

	"github.com/gabkaclassic/metrics/internal/model"
	"github.com/gabkaclassic/metrics/internal/storage"
	"github.com/gabkaclassic/metrics/pkg/metric"
)

type MetricsRepository interface {
	Add(metric models.Metrics) error
	AddAll(metrics *[]models.Metrics) error
	ResetAll(metrics *[]models.Metrics) error
	Reset(metric models.Metrics) error
	Get(metricID string) (*models.Metrics, error)
	GetAll() (*map[string]any, error)
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

func (repository *memoryMetricsRepository) GetAll() (*map[string]any, error) {

	metrics := make(map[string]any, len(repository.storage.Metrics))

	for id, m := range repository.storage.Metrics {
		switch m.MType {
		case string(metric.CounterType):
			metrics[id] = *m.Delta
		case string(metric.GaugeType):
			metrics[id] = *m.Value
		}
	}

	return &metrics, nil
}

func (repository *memoryMetricsRepository) Get(metricID string) (*models.Metrics, error) {

	metric, exists := repository.storage.Metrics[metricID]

	if !exists {
		return nil, fmt.Errorf("metric %s not found", metricID)
	}

	return &metric, nil
}

func (repository *memoryMetricsRepository) updateMetric(metric models.Metrics, updateMetricFunction func(metric models.Metrics) error) error {
	repository.mutex.Lock()
	defer repository.mutex.Unlock()

	if repository.storage.Metrics == nil {
		repository.storage.Metrics = make(map[string]models.Metrics)
	}

	err := updateMetricFunction(metric)

	return err
}
func (repository *memoryMetricsRepository) updateMetrics(metrics *[]models.Metrics, updateMetricsFunction func(metric *[]models.Metrics) error) error {
	repository.mutex.Lock()
	defer repository.mutex.Unlock()

	if repository.storage.Metrics == nil {
		repository.storage.Metrics = make(map[string]models.Metrics)
	}

	err := updateMetricsFunction(metrics)

	return err
}

func (repository *memoryMetricsRepository) Add(metric models.Metrics) error {

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

func (repository *memoryMetricsRepository) AddAll(metrics *[]models.Metrics) error {

	err := repository.updateMetrics(
		metrics,
		func(metrics *[]models.Metrics) error {
			for _, metric := range *metrics {
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

func (repository *memoryMetricsRepository) Reset(metric models.Metrics) error {

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

func (repository *memoryMetricsRepository) ResetAll(metrics *[]models.Metrics) error {

	err := repository.updateMetrics(
		metrics,
		func(metrics *[]models.Metrics) error {
			for _, metric := range *metrics {
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

type dbMetricsRepository struct {
	storage *sql.DB
}

func NewDBMetricsRepository(storage *sql.DB) (MetricsRepository, error) {

	if storage == nil {
		return nil, errors.New("create new metrics repository failed: storage is nil")
	}

	return &dbMetricsRepository{
		storage: storage,
	}, nil
}

func (repository *dbMetricsRepository) GetAll() (*map[string]any, error) {

	rows, err := repository.storage.Query("SELECT id, type, delta, value FROM metric;")

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	metrics := make(map[string]any)

	for rows.Next() {
		var m models.Metrics
		if err = rows.Scan(&m.ID, &m.MType, &m.Delta, &m.Value); err != nil {
			return nil, err
		}
		switch m.MType {
		case string(metric.CounterType):
			metrics[m.ID] = *m.Delta
		case string(metric.GaugeType):
			metrics[m.ID] = *m.Value
		}
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return &metrics, nil
}

func (repository *dbMetricsRepository) Get(metricID string) (*models.Metrics, error) {

	var metric models.Metrics
	err := repository.storage.QueryRow(
		"SELECT id, type, delta, value FROM metric WHERE id = $1",
		metricID).
		Scan(&metric.ID, &metric.MType, &metric.Delta, &metric.Value)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("metric %s not found", metricID)
		}
		return nil, err
	}

	return &metric, nil
}

func (repository *dbMetricsRepository) Add(metric models.Metrics) error {

	tx, err := repository.storage.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	_, err = tx.Exec(
		`INSERT INTO metric (id, type, delta)
		VALUES ($1, 'counter', $2)
		ON CONFLICT (id)
		DO UPDATE SET delta = metric.delta + EXCLUDED.delta;`,
		metric.ID, metric.Delta,
	)

	if err != nil {
		return err
	}

	return tx.Commit()
}

func (repository *dbMetricsRepository) AddAll(metrics *[]models.Metrics) error {

	tx, err := repository.storage.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	ids := make([]string, len(*metrics))
	deltas := make([]int64, len(*metrics))

	for i, metric := range *metrics {
		ids[i] = metric.ID
		deltas[i] = *metric.Delta
	}

	_, err = tx.Exec(`
        INSERT INTO metric (id, type, delta)
        SELECT unnest($1::text[]), 'counter', unnest($2::bigint[])
        ON CONFLICT (id) DO UPDATE 
        SET delta = metric.delta + EXCLUDED.delta
    ;`, pq.Array(ids), pq.Array(deltas))

	if err != nil {
		return err
	}

	return tx.Commit()
}

func (repository *dbMetricsRepository) Reset(metric models.Metrics) error {

	tx, err := repository.storage.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	_, err = tx.Exec(
		`INSERT INTO metric (id, type, value)
		VALUES ($1, 'gauge', $2)
		ON CONFLICT (id)
		DO UPDATE SET value = EXCLUDED.value;`,
		metric.ID, metric.Value,
	)

	if err != nil {
		return err
	}

	return tx.Commit()
}

func (repository *dbMetricsRepository) ResetAll(metrics *[]models.Metrics) error {

	tx, err := repository.storage.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	ids := make([]string, len(*metrics))
	values := make([]float64, len(*metrics))

	for i, metric := range *metrics {
		ids[i] = metric.ID
		values[i] = *metric.Value
	}

	_, err = tx.Exec(`
        INSERT INTO metric (id, type, delta)
        SELECT unnest($1::text[]), 'gauge', unnest($2::bigint[])
        ON CONFLICT (id) DO UPDATE 
        SET value = EXCLUDED.value;
    ;`, pq.Array(ids), pq.Array(values))

	if err != nil {
		return err
	}

	return tx.Commit()
}
