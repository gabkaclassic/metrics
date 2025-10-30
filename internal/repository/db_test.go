package repository

import (
	"database/sql"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gabkaclassic/metrics/internal/model"
	"github.com/gabkaclassic/metrics/pkg/metric"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"regexp"
	"testing"
)

func TestNewDBMetricsRepository(t *testing.T) {
	tests := []struct {
		name        string
		storage     *sql.DB
		expectError bool
	}{
		{
			name:        "valid db connection",
			storage:     &sql.DB{},
			expectError: false,
		},
		{
			name:        "nil storage",
			storage:     nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, err := NewDBMetricsRepository(tt.storage)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, repo)
				assert.Contains(t, err.Error(), "storage is nil")
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, repo)

				dbRepo, ok := repo.(*dbMetricsRepository)
				assert.True(t, ok)
				assert.Equal(t, tt.storage, dbRepo.storage)
			}
		})
	}
}

func TestDBMetricsRepository_GetAll(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo, err := NewDBMetricsRepository(db)
	assert.NoError(t, err)

	tests := []struct {
		name        string
		mockQuery   func()
		expectData  *map[string]any
		expectError bool
	}{
		{
			name: "success mixed metrics",
			mockQuery: func() {
				rows := sqlmock.NewRows([]string{"id", "type", "delta", "value"}).
					AddRow("counter1", string(metric.CounterType), int64(5), nil).
					AddRow("gauge1", string(metric.GaugeType), nil, float64(3.14))
				mock.ExpectQuery("SELECT id, type, delta, value FROM metric;").WillReturnRows(rows)
			},
			expectData: &map[string]any{
				"counter1": int64(5),
				"gauge1":   float64(3.14),
			},
			expectError: false,
		},
		{
			name: "query error",
			mockQuery: func() {
				mock.ExpectQuery("SELECT id, type, delta, value FROM metric;").
					WillReturnError(errors.New("db failure"))
			},
			expectData:  nil,
			expectError: true,
		},
		{
			name: "scan error",
			mockQuery: func() {
				rows := sqlmock.NewRows([]string{"id", "type", "delta", "value"}).
					AddRow(nil, nil, nil, nil)
				mock.ExpectQuery("SELECT id, type, delta, value FROM metric;").WillReturnRows(rows)
			},
			expectData:  nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockQuery()

			result, err := repo.GetAll()

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectData, result)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestDBMetricsRepository_Get(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo, err := NewDBMetricsRepository(db)
	assert.NoError(t, err)

	tests := []struct {
		name        string
		metricID    string
		mockQuery   func()
		expectValue *models.Metrics
		expectError bool
		errorText   string
	}{
		{
			name:     "success gauge metric",
			metricID: "g1",
			mockQuery: func() {
				rows := sqlmock.NewRows([]string{"id", "type", "delta", "value"}).
					AddRow("g1", string(metric.GaugeType), nil, float64(12.34))
				mock.ExpectQuery("SELECT id, type, delta, value FROM metric WHERE id = \\$1").
					WithArgs("g1").WillReturnRows(rows)
			},
			expectValue: &models.Metrics{
				ID:    "g1",
				MType: string(metric.GaugeType),
				Value: floatPtr(12.34),
			},
			expectError: false,
		},
		{
			name:     "success counter metric",
			metricID: "c1",
			mockQuery: func() {
				rows := sqlmock.NewRows([]string{"id", "type", "delta", "value"}).
					AddRow("c1", string(metric.CounterType), int64(7), nil)
				mock.ExpectQuery("SELECT id, type, delta, value FROM metric WHERE id = \\$1").
					WithArgs("c1").WillReturnRows(rows)
			},
			expectValue: &models.Metrics{
				ID:    "c1",
				MType: string(metric.CounterType),
				Delta: intPtr(7),
			},
			expectError: false,
		},
		{
			name:     "metric not found",
			metricID: "missing",
			mockQuery: func() {
				mock.ExpectQuery("SELECT id, type, delta, value FROM metric WHERE id = \\$1").
					WithArgs("missing").WillReturnError(sql.ErrNoRows)
			},
			expectValue: nil,
			expectError: true,
			errorText:   "metric missing not found",
		},
		{
			name:     "query error",
			metricID: "broken",
			mockQuery: func() {
				mock.ExpectQuery("SELECT id, type, delta, value FROM metric WHERE id = \\$1").
					WithArgs("broken").WillReturnError(errors.New("db error"))
			},
			expectValue: nil,
			expectError: true,
			errorText:   "db error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockQuery()

			result, err := repo.Get(tt.metricID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), tt.errorText)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectValue.ID, result.ID)
				assert.Equal(t, tt.expectValue.MType, result.MType)
				assert.Equal(t, tt.expectValue.Delta, result.Delta)
				assert.Equal(t, tt.expectValue.Value, result.Value)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestDBMetricsRepository_Add(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo, err := NewDBMetricsRepository(db)
	assert.NoError(t, err)

	tests := []struct {
		name        string
		metric      models.Metrics
		mockQuery   func()
		expectError bool
		errorText   string
	}{
		{
			name: "success insert new counter",
			metric: models.Metrics{
				ID:    "c1",
				MType: string(metric.CounterType),
				Delta: intPtr(10),
			},
			mockQuery: func() {
				mock.ExpectBegin()
				mock.ExpectExec(`INSERT INTO metric \(id, type, delta\)`).
					WithArgs("c1", intPtr(10)).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			expectError: false,
		},
		{
			name: "begin transaction error",
			metric: models.Metrics{
				ID:    "c2",
				MType: string(metric.CounterType),
				Delta: intPtr(5),
			},
			mockQuery: func() {
				mock.ExpectBegin().WillReturnError(errors.New("tx begin failed"))
			},
			expectError: true,
			errorText:   "tx begin failed",
		},
		{
			name: "exec error",
			metric: models.Metrics{
				ID:    "c3",
				MType: string(metric.CounterType),
				Delta: intPtr(15),
			},
			mockQuery: func() {
				mock.ExpectBegin()
				mock.ExpectExec(`INSERT INTO metric \(id, type, delta\)`).
					WithArgs("c3", intPtr(15)).
					WillReturnError(errors.New("insert failed"))
				mock.ExpectRollback()
			},
			expectError: true,
			errorText:   "insert failed",
		},
		{
			name: "commit error",
			metric: models.Metrics{
				ID:    "c4",
				MType: string(metric.CounterType),
				Delta: intPtr(20),
			},
			mockQuery: func() {
				mock.ExpectBegin()
				mock.ExpectExec(`INSERT INTO metric \(id, type, delta\)`).
					WithArgs("c4", intPtr(20)).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit().WillReturnError(errors.New("commit failed"))
			},
			expectError: true,
			errorText:   "commit failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockQuery()

			err := repo.Add(tt.metric)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorText)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestDBMetricsRepository_Reset(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo, err := NewDBMetricsRepository(db)
	assert.NoError(t, err)

	tests := []struct {
		name        string
		metric      models.Metrics
		mockQuery   func()
		expectError bool
		errorText   string
	}{
		{
			name: "success gauge metric reset",
			metric: models.Metrics{
				ID:    "g1",
				MType: string(metric.GaugeType),
				Value: floatPtr(42.42),
			},
			mockQuery: func() {
				mock.ExpectBegin()
				mock.ExpectExec(`INSERT INTO metric \(id, type, value\)`).
					WithArgs("g1", floatPtr(42.42)).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			expectError: false,
		},
		{
			name: "begin transaction error",
			metric: models.Metrics{
				ID:    "g2",
				MType: string(metric.GaugeType),
				Value: floatPtr(3.14),
			},
			mockQuery: func() {
				mock.ExpectBegin().WillReturnError(errors.New("tx begin failed"))
			},
			expectError: true,
			errorText:   "tx begin failed",
		},
		{
			name: "exec error",
			metric: models.Metrics{
				ID:    "g3",
				MType: string(metric.GaugeType),
				Value: floatPtr(1.23),
			},
			mockQuery: func() {
				mock.ExpectBegin()
				mock.ExpectExec(`INSERT INTO metric \(id, type, value\)`).
					WithArgs("g3", floatPtr(1.23)).
					WillReturnError(errors.New("insert failed"))
				mock.ExpectRollback()
			},
			expectError: true,
			errorText:   "insert failed",
		},
		{
			name: "commit error",
			metric: models.Metrics{
				ID:    "g4",
				MType: string(metric.GaugeType),
				Value: floatPtr(99.9),
			},
			mockQuery: func() {
				mock.ExpectBegin()
				mock.ExpectExec(`INSERT INTO metric \(id, type, value\)`).
					WithArgs("g4", floatPtr(99.9)).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit().WillReturnError(errors.New("commit failed"))
			},
			expectError: true,
			errorText:   "commit failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockQuery()

			err := repo.Reset(tt.metric)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorText)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestDBMetricsRepository_AddAll(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo, err := NewDBMetricsRepository(db)
	assert.NoError(t, err)

	tests := []struct {
		name        string
		metrics     *[]models.Metrics
		mockQuery   func()
		expectError bool
	}{
		{
			name: "success single counter",
			metrics: &[]models.Metrics{
				{ID: "c1", MType: models.Counter, Delta: intPtr(10)},
			},
			mockQuery: func() {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(`
        INSERT INTO metric (id, type, delta)
        SELECT unnest($1::text[]), 'counter', unnest($2::bigint[])
        ON CONFLICT (id) DO UPDATE 
        SET delta = metric.delta + EXCLUDED.delta
    ;`)).
					WithArgs(pq.Array([]string{"c1"}), pq.Array([]int64{10})).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit()
			},
			expectError: false,
		},
		{
			name: "success multiple counters",
			metrics: &[]models.Metrics{
				{ID: "c1", MType: models.Counter, Delta: intPtr(10)},
				{ID: "c2", MType: models.Counter, Delta: intPtr(5)},
				{ID: "c1", MType: models.Counter, Delta: intPtr(3)},
			},
			mockQuery: func() {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(`
        INSERT INTO metric (id, type, delta)
        SELECT unnest($1::text[]), 'counter', unnest($2::bigint[])
        ON CONFLICT (id) DO UPDATE 
        SET delta = metric.delta + EXCLUDED.delta
    ;`)).
					WithArgs(pq.Array([]string{"c1", "c2", "c1"}), pq.Array([]int64{10, 5, 3})).
					WillReturnResult(sqlmock.NewResult(0, 3))
				mock.ExpectCommit()
			},
			expectError: false,
		},
		{
			name:    "empty metrics",
			metrics: &[]models.Metrics{},
			mockQuery: func() {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(`
        INSERT INTO metric (id, type, delta)
        SELECT unnest($1::text[]), 'counter', unnest($2::bigint[])
        ON CONFLICT (id) DO UPDATE 
        SET delta = metric.delta + EXCLUDED.delta
    ;`)).
					WithArgs(pq.Array([]string{}), pq.Array([]int64{})).
					WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectCommit()
			},
			expectError: false,
		},
		{
			name: "begin transaction error",
			metrics: &[]models.Metrics{
				{ID: "c1", MType: models.Counter, Delta: intPtr(10)},
			},
			mockQuery: func() {
				mock.ExpectBegin().WillReturnError(errors.New("begin error"))
			},
			expectError: true,
		},
		{
			name: "exec query error",
			metrics: &[]models.Metrics{
				{ID: "c1", MType: models.Counter, Delta: intPtr(10)},
			},
			mockQuery: func() {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(`
        INSERT INTO metric (id, type, delta)
        SELECT unnest($1::text[]), 'counter', unnest($2::bigint[])
        ON CONFLICT (id) DO UPDATE 
        SET delta = metric.delta + EXCLUDED.delta
    ;`)).
					WithArgs(pq.Array([]string{"c1"}), pq.Array([]int64{10})).
					WillReturnError(errors.New("exec error"))
				mock.ExpectRollback()
			},
			expectError: true,
		},
		{
			name: "commit error",
			metrics: &[]models.Metrics{
				{ID: "c1", MType: models.Counter, Delta: intPtr(10)},
			},
			mockQuery: func() {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(`
        INSERT INTO metric (id, type, delta)
        SELECT unnest($1::text[]), 'counter', unnest($2::bigint[])
        ON CONFLICT (id) DO UPDATE 
        SET delta = metric.delta + EXCLUDED.delta
    ;`)).
					WithArgs(pq.Array([]string{"c1"}), pq.Array([]int64{10})).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit().WillReturnError(errors.New("commit error"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockQuery()

			err := repo.AddAll(tt.metrics)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestDBMetricsRepository_ResetAll(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo, err := NewDBMetricsRepository(db)
	assert.NoError(t, err)

	tests := []struct {
		name        string
		metrics     *[]models.Metrics
		mockQuery   func()
		expectError bool
	}{
		{
			name: "success single gauge",
			metrics: &[]models.Metrics{
				{ID: "g1", MType: models.Gauge, Value: floatPtr(3.14)},
			},
			mockQuery: func() {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(`
        INSERT INTO metric (id, type, value)
        SELECT unnest($1::text[]), 'gauge', unnest($2::float8[])
        ON CONFLICT (id) DO UPDATE 
        SET value = EXCLUDED.value;
    ;`)).
					WithArgs(pq.Array([]string{"g1"}), pq.Array([]float64{3.14})).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit()
			},
			expectError: false,
		},
		{
			name: "success multiple gauges",
			metrics: &[]models.Metrics{
				{ID: "g1", MType: models.Gauge, Value: floatPtr(3.14)},
				{ID: "g2", MType: models.Gauge, Value: floatPtr(2.71)},
				{ID: "g3", MType: models.Gauge, Value: floatPtr(1.41)},
			},
			mockQuery: func() {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(`
        INSERT INTO metric (id, type, value)
        SELECT unnest($1::text[]), 'gauge', unnest($2::float8[])
        ON CONFLICT (id) DO UPDATE 
        SET value = EXCLUDED.value;
    ;`)).
					WithArgs(pq.Array([]string{"g1", "g2", "g3"}), pq.Array([]float64{3.14, 2.71, 1.41})).
					WillReturnResult(sqlmock.NewResult(0, 3))
				mock.ExpectCommit()
			},
			expectError: false,
		},
		{
			name:    "empty metrics",
			metrics: &[]models.Metrics{},
			mockQuery: func() {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(`
        INSERT INTO metric (id, type, value)
        SELECT unnest($1::text[]), 'gauge', unnest($2::float8[])
        ON CONFLICT (id) DO UPDATE 
        SET value = EXCLUDED.value;
    ;`)).
					WithArgs(pq.Array([]string{}), pq.Array([]float64{})).
					WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectCommit()
			},
			expectError: false,
		},
		{
			name: "begin transaction error",
			metrics: &[]models.Metrics{
				{ID: "g1", MType: models.Gauge, Value: floatPtr(3.14)},
			},
			mockQuery: func() {
				mock.ExpectBegin().WillReturnError(errors.New("begin error"))
			},
			expectError: true,
		},
		{
			name: "exec query error",
			metrics: &[]models.Metrics{
				{ID: "g1", MType: models.Gauge, Value: floatPtr(3.14)},
			},
			mockQuery: func() {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(`
        INSERT INTO metric (id, type, value)
        SELECT unnest($1::text[]), 'gauge', unnest($2::float8[])
        ON CONFLICT (id) DO UPDATE 
        SET value = EXCLUDED.value;
    ;`)).
					WithArgs(pq.Array([]string{"g1"}), pq.Array([]float64{3.14})).
					WillReturnError(errors.New("exec error"))
				mock.ExpectRollback()
			},
			expectError: true,
		},
		{
			name: "commit error",
			metrics: &[]models.Metrics{
				{ID: "g1", MType: models.Gauge, Value: floatPtr(3.14)},
			},
			mockQuery: func() {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(`
        INSERT INTO metric (id, type, value)
        SELECT unnest($1::text[]), 'gauge', unnest($2::float8[])
        ON CONFLICT (id) DO UPDATE 
        SET value = EXCLUDED.value;
    ;`)).
					WithArgs(pq.Array([]string{"g1"}), pq.Array([]float64{3.14})).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit().WillReturnError(errors.New("commit error"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockQuery()

			err := repo.ResetAll(tt.metrics)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestDBMetricsRepository_GetAllMetrics(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo, err := NewDBMetricsRepository(db)
	assert.NoError(t, err)

	tests := []struct {
		name        string
		mockQuery   func()
		expectData  *[]models.Metrics
		expectError bool
	}{
		{
			name: "success mixed metrics",
			mockQuery: func() {
				rows := sqlmock.NewRows([]string{"id", "type", "delta", "value"}).
					AddRow("counter1", string(models.Counter), int64(5), nil).
					AddRow("gauge1", string(models.Gauge), nil, float64(3.14))
				mock.ExpectQuery("SELECT id, type, delta, value FROM metric;").WillReturnRows(rows)
			},
			expectData: &[]models.Metrics{
				{ID: "counter1", MType: models.Counter, Delta: intPtr(5), Value: nil},
				{ID: "gauge1", MType: models.Gauge, Delta: nil, Value: floatPtr(3.14)},
			},
			expectError: false,
		},
		{
			name: "success only counters",
			mockQuery: func() {
				rows := sqlmock.NewRows([]string{"id", "type", "delta", "value"}).
					AddRow("counter1", string(models.Counter), int64(10), nil).
					AddRow("counter2", string(models.Counter), int64(20), nil)
				mock.ExpectQuery("SELECT id, type, delta, value FROM metric;").WillReturnRows(rows)
			},
			expectData: &[]models.Metrics{
				{ID: "counter1", MType: models.Counter, Delta: intPtr(10), Value: nil},
				{ID: "counter2", MType: models.Counter, Delta: intPtr(20), Value: nil},
			},
			expectError: false,
		},
		{
			name: "success only gauges",
			mockQuery: func() {
				rows := sqlmock.NewRows([]string{"id", "type", "delta", "value"}).
					AddRow("gauge1", string(models.Gauge), nil, float64(1.1)).
					AddRow("gauge2", string(models.Gauge), nil, float64(2.2))
				mock.ExpectQuery("SELECT id, type, delta, value FROM metric;").WillReturnRows(rows)
			},
			expectData: &[]models.Metrics{
				{ID: "gauge1", MType: models.Gauge, Delta: nil, Value: floatPtr(1.1)},
				{ID: "gauge2", MType: models.Gauge, Delta: nil, Value: floatPtr(2.2)},
			},
			expectError: false,
		},
		{
			name: "empty result",
			mockQuery: func() {
				rows := sqlmock.NewRows([]string{"id", "type", "delta", "value"})
				mock.ExpectQuery("SELECT id, type, delta, value FROM metric;").WillReturnRows(rows)
			},
			expectData:  &[]models.Metrics{},
			expectError: false,
		},
		{
			name: "query error",
			mockQuery: func() {
				mock.ExpectQuery("SELECT id, type, delta, value FROM metric;").
					WillReturnError(errors.New("db failure"))
			},
			expectData:  nil,
			expectError: true,
		},
		{
			name: "scan error",
			mockQuery: func() {
				rows := sqlmock.NewRows([]string{"id", "type", "delta", "value"}).
					AddRow(nil, nil, nil, nil)
				mock.ExpectQuery("SELECT id, type, delta, value FROM metric;").WillReturnRows(rows)
			},
			expectData:  nil,
			expectError: true,
		},
		{
			name: "rows error",
			mockQuery: func() {
				rows := sqlmock.NewRows([]string{"id", "type", "delta", "value"}).
					AddRow("counter1", string(models.Counter), int64(5), nil).
					RowError(0, errors.New("row error"))
				mock.ExpectQuery("SELECT id, type, delta, value FROM metric;").WillReturnRows(rows)
			},
			expectData:  nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockQuery()

			result, err := repo.GetAllMetrics()

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, len(*tt.expectData), len(*result))
				for i, expectedMetric := range *tt.expectData {
					actualMetric := (*result)[i]
					assert.Equal(t, expectedMetric.ID, actualMetric.ID)
					assert.Equal(t, expectedMetric.MType, actualMetric.MType)
					if expectedMetric.Delta != nil {
						assert.Equal(t, *expectedMetric.Delta, *actualMetric.Delta)
					} else {
						assert.Nil(t, actualMetric.Delta)
					}
					if expectedMetric.Value != nil {
						assert.Equal(t, *expectedMetric.Value, *actualMetric.Value)
					} else {
						assert.Nil(t, actualMetric.Value)
					}
				}
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
