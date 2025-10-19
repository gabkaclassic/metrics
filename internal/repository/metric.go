package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"sync"
	"time"

	"github.com/gabkaclassic/metrics/internal/model"
	"github.com/gabkaclassic/metrics/internal/storage"
	"github.com/gabkaclassic/metrics/pkg/metric"
)

const (
	retriesAmount int           = 3
	retryDelay    time.Duration = 1 * time.Second
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
	var metrics *map[string]any
	err := repository.executeWithRetry(func() error {
		rows, err := repository.storage.Query("SELECT id, type, delta, value FROM metric;")
		if err != nil {
			return err
		}
		defer rows.Close()

		currentMetrics := make(map[string]any)
		for rows.Next() {
			var m models.Metrics
			if err = rows.Scan(&m.ID, &m.MType, &m.Delta, &m.Value); err != nil {
				return err
			}
			switch m.MType {
			case string(metric.CounterType):
				currentMetrics[m.ID] = *m.Delta
			case string(metric.GaugeType):
				currentMetrics[m.ID] = *m.Value
			}
		}

		if err = rows.Err(); err != nil {
			return err
		}

		metrics = &currentMetrics
		return nil
	})

	if err != nil {
		return nil, err
	}
	return metrics, nil
}

func (repository *dbMetricsRepository) Get(metricID string) (*models.Metrics, error) {
	var metric *models.Metrics
	err := repository.executeWithRetry(func() error {
		m := models.Metrics{}
		err := repository.storage.QueryRow(
			"SELECT id, type, delta, value FROM metric WHERE id = $1",
			metricID).
			Scan(&m.ID, &m.MType, &m.Delta, &m.Value)

		if err != nil {
			if err == sql.ErrNoRows {
				return fmt.Errorf("metric %s not found", metricID)
			}
			return err
		}
		metric = &m
		return nil
	})

	if err != nil {
		return nil, err
	}
	return metric, nil
}

func (repository *dbMetricsRepository) Add(metric models.Metrics) error {
	return repository.executeWithRetry(func() error {
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
	})
}

func (repository *dbMetricsRepository) AddAll(metrics *[]models.Metrics) error {
	return repository.executeWithRetry(func() error {
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
	})
}

func (repository *dbMetricsRepository) Reset(metric models.Metrics) error {
	return repository.executeWithRetry(func() error {
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
	})
}

func (repository *dbMetricsRepository) ResetAll(metrics *[]models.Metrics) error {
	return repository.executeWithRetry(func() error {
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
			INSERT INTO metric (id, type, value)
			SELECT unnest($1::text[]), 'gauge', unnest($2::float8[])
			ON CONFLICT (id) DO UPDATE 
			SET value = EXCLUDED.value;
		;`, pq.Array(ids), pq.Array(values))

		if err != nil {
			return err
		}

		return tx.Commit()
	})
}

func isRetryableError(err error) bool {
	if pqErr, ok := err.(*pq.Error); ok {
		switch pqErr.Code {
		case "08000", // connection_exception
			"08003", // connection_does_not_exist
			"08006", // connection_failure
			"08001", // sqlclient_unable_to_establish_sqlconnection
			"08004", // sqlserver_rejected_establishment_of_sqlconnection
			"08007", // transaction_resolution_unknown
			"40001", // serialization_failure
			"40P01", // deadlock_detected
			"55006", // object_in_use
			"55P03": // lock_not_available
			return true
		}
	}
	return false
}

func (repository *dbMetricsRepository) executeWithRetry(operation func() error) error {
	var lastErr error
	currentRetryDelay := retryDelay
	for i := 0; i < retriesAmount; i++ {
		lastErr = operation()
		if lastErr == nil {
			return nil
		}

		if !isRetryableError(lastErr) {
			return lastErr
		}

		if i < retriesAmount-1 {
			time.Sleep(currentRetryDelay)
			currentRetryDelay *= 2
		}
	}
	return lastErr
}
