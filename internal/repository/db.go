package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"time"

	"github.com/gabkaclassic/metrics/internal/model"
	"github.com/gabkaclassic/metrics/pkg/metric"
)

const (
	retriesAmount int           = 3
	retryDelay    time.Duration = 1 * time.Second
)

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

func (repository *dbMetricsRepository) GetAllMetrics(ctx context.Context) ([]models.Metrics, error) {
	metrics := make([]models.Metrics, 0)
	err := repository.executeWithRetry(func() error {
		rows, err := repository.storage.QueryContext(ctx, "SELECT id, type, delta, value FROM metric;")
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var m models.Metrics
			if err = rows.Scan(&m.ID, &m.MType, &m.Delta, &m.Value); err != nil {
				return err
			}
			metrics = append(metrics, m)
		}

		if err = rows.Err(); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	return metrics, nil
}

func (repository *dbMetricsRepository) GetAll(ctx context.Context) (map[string]any, error) {
	var metrics map[string]any
	err := repository.executeWithRetry(func() error {
		rows, err := repository.storage.QueryContext(ctx, "SELECT id, type, delta, value FROM metric;")
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

		metrics = currentMetrics
		return nil
	})

	if err != nil {
		return nil, err
	}
	return metrics, nil
}

func (repository *dbMetricsRepository) Get(ctx context.Context, metricID string) (*models.Metrics, error) {
	var metric models.Metrics
	err := repository.executeWithRetry(func() error {
		m := models.Metrics{}
		err := repository.storage.QueryRowContext(
			ctx,
			"SELECT id, type, delta, value FROM metric WHERE id = $1",
			metricID,
		).
			Scan(&m.ID, &m.MType, &m.Delta, &m.Value)

		if err != nil {
			if err == sql.ErrNoRows {
				return fmt.Errorf("metric %s not found", metricID)
			}
			return err
		}
		metric = m
		return nil
	})

	if err != nil {
		return nil, err
	}
	return &metric, nil
}

func (repository *dbMetricsRepository) Add(ctx context.Context, metric models.Metrics) error {
	return repository.executeWithRetry(func() error {
		tx, err := repository.storage.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()

		_, err = tx.ExecContext(
			ctx,
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

func (repository *dbMetricsRepository) AddAll(ctx context.Context, metrics []models.Metrics) error {
	return repository.executeWithRetry(func() error {
		tx, err := repository.storage.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()

		ids := make([]string, len(metrics))
		deltas := make([]int64, len(metrics))

		for i, metric := range metrics {
			ids[i] = metric.ID
			deltas[i] = *metric.Delta
		}

		_, err = tx.ExecContext(
			ctx,
			`
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

func (repository *dbMetricsRepository) Reset(ctx context.Context, metric models.Metrics) error {
	return repository.executeWithRetry(func() error {
		tx, err := repository.storage.BeginTx(ctx, nil)
		if err != nil {
			return err
		}

		defer tx.Rollback()

		_, err = tx.ExecContext(
			ctx,
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

func (repository *dbMetricsRepository) ResetAll(ctx context.Context, metrics []models.Metrics) error {
	return repository.executeWithRetry(func() error {
		tx, err := repository.storage.BeginTx(ctx, nil)
		if err != nil {
			return err
		}

		defer tx.Rollback()

		ids := make([]string, len(metrics))
		values := make([]float64, len(metrics))

		for i, metric := range metrics {
			ids[i] = metric.ID
			values[i] = *metric.Value
		}

		_, err = tx.ExecContext(
			ctx,
			`
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
			timer := time.NewTimer(currentRetryDelay)
			<-timer.C
			currentRetryDelay *= 2
		}
	}
	return lastErr
}
