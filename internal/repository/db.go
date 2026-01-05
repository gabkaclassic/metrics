package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	models "github.com/gabkaclassic/metrics/internal/model"
	"github.com/gabkaclassic/metrics/internal/storage"
	"github.com/gabkaclassic/metrics/pkg/metric"
)

const (
	// retriesAmount defines the maximum number of retry attempts for database operations.
	retriesAmount int = 3

	// retryDelay defines the initial delay between retry attempts.
	// Uses exponential backoff: delay doubles after each retry.
	retryDelay time.Duration = 1 * time.Second
)

// dbMetricsRepository implements MetricsRepository using PostgreSQL database.
// Provides persistent storage with ACID compliance and transaction support.
// generate:reset
type dbMetricsRepository struct {
	storage storage.DB
}

// NewDBMetricsRepository creates a new PostgreSQL-based metrics repository.
//
// storage: Established SQL database connection (typically PostgreSQL)
//
// Returns:
//   - MetricsRepository: Ready-to-use repository instance
//   - error: If storage connection is nil
//
// The repository automatically retries operations on transient database errors.
func NewDBMetricsRepository(s storage.DB) (MetricsRepository, error) {
	if s == nil {
		return nil, errors.New("create new metrics repository failed: storage is nil")
	}

	return &dbMetricsRepository{
		storage: s,
	}, nil
}

// GetAllMetrics retrieves all metrics from the database.
// Returns metrics in their complete structure including type and values.
func (repository *dbMetricsRepository) GetAllMetrics(ctx context.Context) ([]models.Metrics, error) {
	metrics := make([]models.Metrics, 0)
	err := repository.executeWithRetry(func() error {
		rows, err := repository.storage.Query(ctx, "SELECT id, type, delta, value FROM metric;")
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var m models.Metrics
			var delta pgtype.Int8
			var value pgtype.Float8
			if err = rows.Scan(&m.ID, &m.MType, &delta, &value); err != nil {
				return err
			}
			switch m.MType {
			case string(metric.CounterType):
				m.Delta = &delta.Int64
			case string(metric.GaugeType):
				m.Value = &value.Float64
			default:
				return fmt.Errorf("invalid metric type: %s", m.MType)
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

// GetAll returns all metrics as a map of metric ID to value.
// Counter metrics are returned as int64, gauge metrics as float64.
// Performs a single database query with automatic retry on failure.
func (repository *dbMetricsRepository) GetAll(ctx context.Context) (map[string]any, error) {
	var metrics map[string]any
	err := repository.executeWithRetry(func() error {
		rows, err := repository.storage.Query(ctx, "SELECT id, type, delta, value FROM metric;")
		if err != nil {
			return err
		}
		defer rows.Close()

		currentMetrics := make(map[string]any)
		for rows.Next() {
			var m models.Metrics
			var delta pgtype.Int8
			var value pgtype.Float8

			if err = rows.Scan(&m.ID, &m.MType, &delta, &value); err != nil {
				return err
			}
			switch m.MType {
			case string(metric.CounterType):
				currentMetrics[m.ID] = delta.Int64
			case string(metric.GaugeType):
				currentMetrics[m.ID] = value.Float64
			default:
				return fmt.Errorf("invalid metric type: %s", m.MType)
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

// Get retrieves a single metric by its ID from the database.
// Returns sql.ErrNoRows wrapped in a descriptive error if metric not found.
func (repository *dbMetricsRepository) Get(ctx context.Context, metricID string) (*models.Metrics, error) {
	var result models.Metrics
	err := repository.executeWithRetry(func() error {
		m := models.Metrics{}
		var delta pgtype.Int8
		var value pgtype.Float8
		err := repository.storage.QueryRow(
			ctx,
			"SELECT id, type, delta, value FROM metric WHERE id = $1",
			metricID,
		).
			Scan(&m.ID, &m.MType, &delta, &value)

		if err != nil {
			if err == sql.ErrNoRows {
				return fmt.Errorf("metric %s not found", metricID)
			}
			return err
		}
		switch m.MType {
		case string(metric.CounterType):
			m.Delta = &delta.Int64
		case string(metric.GaugeType):
			m.Value = &value.Float64
		default:
			return fmt.Errorf("invalid metric type: %s", m.MType)
		}
		result = m
		return nil
	})

	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Add increments a counter metric in the database.
// Uses UPSERT pattern: inserts new counter or adds delta to existing one.
// Executes within a transaction with automatic rollback on error.
func (repository *dbMetricsRepository) Add(ctx context.Context, metric models.Metrics) error {
	return repository.executeWithRetry(func() error {
		tx, err := repository.storage.Begin(ctx)
		if err != nil {
			return err
		}
		defer tx.Rollback(ctx)

		_, err = tx.Exec(
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

		return tx.Commit(ctx)
	})
}

// AddAll performs batch addition of counter metrics.
// Uses PostgreSQL array operations for efficient bulk UPSERT.
// More performant than multiple individual Add calls.
func (repository *dbMetricsRepository) AddAll(ctx context.Context, metrics []models.Metrics) error {
	return repository.executeWithRetry(func() error {
		tx, err := repository.storage.Begin(ctx)
		if err != nil {
			return err
		}
		defer tx.Rollback(ctx)

		ids := make([]string, len(metrics))
		deltas := make([]int64, len(metrics))

		for i, metric := range metrics {
			ids[i] = metric.ID
			deltas[i] = *metric.Delta
		}

		_, err = tx.Exec(
			ctx,
			`
			INSERT INTO metric (id, type, delta)
			SELECT unnest($1::text[]), 'counter', unnest($2::bigint[])
			ON CONFLICT (id) DO UPDATE
			SET delta = metric.delta + EXCLUDED.delta
			`,
			ids,
			deltas,
		)
		if err != nil {
			return err
		}

		return tx.Commit(ctx)
	})
}

// Reset sets a gauge metric to a specific value in the database.
// Uses UPSERT pattern: inserts new gauge or updates existing value.
// Executes within a transaction with automatic rollback on error.
func (repository *dbMetricsRepository) ResetOne(ctx context.Context, metric models.Metrics) error {
	return repository.executeWithRetry(func() error {
		tx, err := repository.storage.Begin(ctx)
		if err != nil {
			return err
		}

		defer tx.Rollback(ctx)

		_, err = tx.Exec(
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

		return tx.Commit(ctx)
	})
}

// ResetAll performs batch reset of gauge metrics.
// Uses PostgreSQL array operations for efficient bulk UPSERT.
// Updates existing values or inserts new metrics in a single operation.
func (repository *dbMetricsRepository) ResetAll(ctx context.Context, metrics []models.Metrics) error {
	return repository.executeWithRetry(func() error {
		tx, err := repository.storage.Begin(ctx)
		if err != nil {
			return err
		}

		defer tx.Rollback(ctx)

		ids := make([]string, len(metrics))
		values := make([]float64, len(metrics))

		for i, metric := range metrics {
			ids[i] = metric.ID
			values[i] = *metric.Value
		}

		_, err = tx.Exec(
			ctx,
			`
			INSERT INTO metric (id, type, value)
			SELECT unnest($1::text[]), 'gauge', unnest($2::float8[])
			ON CONFLICT (id) DO UPDATE 
			SET value = EXCLUDED.value;
		;`, ids, values)

		if err != nil {
			return err
		}

		return tx.Commit(ctx)
	})
}

// isRetryableError determines if a database error is transient and safe to retry.
// Checks PostgreSQL error codes for connection issues, deadlocks, and serialization failures.
func isRetryableError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "08000",
			"08003",
			"08006",
			"08001",
			"08004",
			"08007",
			"40001",
			"40P01",
			"55006",
			"55P03":
			return true
		}
	}
	return false
}

// executeWithRetry executes a database operation with automatic retry logic.
// Implements exponential backoff for retryable errors.
// Returns the last error if all retries fail.
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
