package repository

import (
	"database/sql"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gabkaclassic/metrics/internal/model"
	"github.com/gabkaclassic/metrics/internal/storage"
	"github.com/gabkaclassic/metrics/pkg/metric"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"regexp"
	"sync"
	"testing"
)

func TestNewMemoryMetricsRepository(t *testing.T) {
	tests := []struct {
		name        string
		storage     *storage.MemStorage
		expectError bool
	}{
		{
			name:        "valid storage",
			storage:     storage.NewMemStorage(),
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
			repo, err := NewMemoryMetricsRepository(tt.storage, &sync.RWMutex{})

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, repo)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, repo)
			}
		})
	}
}

func TestMetricsRepository_Get(t *testing.T) {
	st := storage.NewMemStorage()
	st.Metrics["existing"] = models.Metrics{ID: "existing", Value: floatPtr(42)}

	repo, err := NewMemoryMetricsRepository(st, &sync.RWMutex{})

	assert.NoError(t, err)

	tests := []struct {
		name        string
		metricID    string
		expectValue *models.Metrics
		expectError bool
	}{
		{
			name:        "metric exists",
			metricID:    "existing",
			expectValue: &models.Metrics{ID: "existing", Value: floatPtr(42)},
			expectError: false,
		},
		{
			name:        "metric does not exist",
			metricID:    "missing",
			expectValue: nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.Get(tt.metricID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectValue, result)
			}
		})
	}
}

func TestMetricsRepository_updateMetric(t *testing.T) {
	storage := storage.NewMemStorage()
	repo := &memoryMetricsRepository{storage: storage, mutex: &sync.RWMutex{}}

	tests := []struct {
		name        string
		metric      models.Metrics
		updateFunc  func(metric models.Metrics) error
		expectError bool
	}{
		{
			name:   "successful update",
			metric: models.Metrics{ID: "m1", Value: floatPtr(10)},
			updateFunc: func(metric models.Metrics) error {
				repo.storage.Metrics[metric.ID] = metric
				return nil
			},
			expectError: false,
		},
		{
			name:   "update returns error",
			metric: models.Metrics{ID: "m2", Value: floatPtr(20)},
			updateFunc: func(metric models.Metrics) error {
				return errors.New("update failed")
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.updateMetric(tt.metric, tt.updateFunc)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				stored, _ := repo.Get(tt.metric.ID)
				assert.Equal(t, tt.metric, *stored)
			}
		})
	}
}

func TestMemoryMetricsRepository_updateMetrics(t *testing.T) {
	repo := &memoryMetricsRepository{
		storage: &storage.MemStorage{
			Metrics: make(map[string]models.Metrics),
		},
		mutex: &sync.RWMutex{},
	}

	tests := []struct {
		name            string
		initialStorage  map[string]models.Metrics
		metrics         *[]models.Metrics
		updateFn        func(metric *[]models.Metrics) error
		expectedError   bool
		expectedStorage map[string]models.Metrics
	}{
		{
			name:           "successful update",
			initialStorage: map[string]models.Metrics{},
			metrics: &[]models.Metrics{
				{ID: "test1", MType: models.Gauge, Value: floatPtr(1.0)},
			},
			updateFn: func(metrics *[]models.Metrics) error {
				return nil
			},
			expectedError:   false,
			expectedStorage: map[string]models.Metrics{},
		},
		{
			name:           "update function error",
			initialStorage: map[string]models.Metrics{},
			metrics: &[]models.Metrics{
				{ID: "test1", MType: models.Gauge, Value: floatPtr(1.0)},
			},
			updateFn: func(metrics *[]models.Metrics) error {
				return errors.New("update error")
			},
			expectedError:   true,
			expectedStorage: map[string]models.Metrics{},
		},
		{
			name:           "initialize nil storage",
			initialStorage: nil,
			metrics: &[]models.Metrics{
				{ID: "test1", MType: models.Gauge, Value: floatPtr(1.0)},
			},
			updateFn: func(metrics *[]models.Metrics) error {
				return nil
			},
			expectedError:   false,
			expectedStorage: map[string]models.Metrics{},
		},
		{
			name:           "empty metrics",
			initialStorage: map[string]models.Metrics{},
			metrics:        &[]models.Metrics{},
			updateFn: func(metrics *[]models.Metrics) error {
				return nil
			},
			expectedError:   false,
			expectedStorage: map[string]models.Metrics{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo.storage.Metrics = tt.initialStorage

			err := repo.updateMetrics(tt.metrics, tt.updateFn)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.NotNil(t, repo.storage.Metrics)
			assert.Equal(t, tt.expectedStorage, repo.storage.Metrics)
		})
	}
}

func TestMetricsRepository_Add(t *testing.T) {
	tests := []struct {
		name           string
		initialMetrics []models.Metrics
		addMetric      models.Metrics
		expectedMetric models.Metrics
	}{
		{
			name:           "add new metric",
			initialMetrics: nil,
			addMetric:      models.Metrics{ID: "m1", Delta: intPtr(5), MType: "counter"},
			expectedMetric: models.Metrics{ID: "m1", Delta: intPtr(5), MType: "counter"},
		},
		{
			name: "update existing metric",
			initialMetrics: []models.Metrics{
				{ID: "m2", Delta: intPtr(3), MType: "counter"},
			},
			addMetric:      models.Metrics{ID: "m2", Delta: intPtr(2), MType: "counter"},
			expectedMetric: models.Metrics{ID: "m2", Delta: intPtr(5), MType: "counter"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			memStorage := storage.NewMemStorage()
			repo, err := NewMemoryMetricsRepository(memStorage, &sync.RWMutex{})
			assert.NoError(t, err)

			for _, m := range tt.initialMetrics {
				err := repo.Add(m)
				assert.NoError(t, err)
			}

			err = repo.Add(tt.addMetric)
			assert.NoError(t, err)

			result, err := repo.Get(tt.addMetric.ID)
			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, *tt.expectedMetric.Delta, *result.Delta)
		})
	}
}

func TestMemoryMetricsRepository_AddAll(t *testing.T) {
	tests := []struct {
		name            string
		initialStorage  map[string]models.Metrics
		metrics         *[]models.Metrics
		expectedStorage map[string]models.Metrics
		expectedError   bool
	}{
		{
			name:           "add new counters",
			initialStorage: map[string]models.Metrics{},
			metrics: &[]models.Metrics{
				{ID: "c1", MType: models.Counter, Delta: intPtr(10)},
				{ID: "c2", MType: models.Counter, Delta: intPtr(5)},
			},
			expectedStorage: map[string]models.Metrics{
				"c1": {ID: "c1", MType: models.Counter, Delta: intPtr(10)},
				"c2": {ID: "c2", MType: models.Counter, Delta: intPtr(5)},
			},
			expectedError: false,
		},
		{
			name: "increment existing counters",
			initialStorage: map[string]models.Metrics{
				"c1": {ID: "c1", MType: models.Counter, Delta: intPtr(10)},
				"c2": {ID: "c2", MType: models.Counter, Delta: intPtr(5)},
			},
			metrics: &[]models.Metrics{
				{ID: "c1", MType: models.Counter, Delta: intPtr(3)},
				{ID: "c2", MType: models.Counter, Delta: intPtr(7)},
				{ID: "c3", MType: models.Counter, Delta: intPtr(1)},
			},
			expectedStorage: map[string]models.Metrics{
				"c1": {ID: "c1", MType: models.Counter, Delta: intPtr(13)},
				"c2": {ID: "c2", MType: models.Counter, Delta: intPtr(12)},
				"c3": {ID: "c3", MType: models.Counter, Delta: intPtr(1)},
			},
			expectedError: false,
		},
		{
			name: "empty metrics",
			initialStorage: map[string]models.Metrics{
				"c1": {ID: "c1", MType: models.Counter, Delta: intPtr(10)},
			},
			metrics: &[]models.Metrics{},
			expectedStorage: map[string]models.Metrics{
				"c1": {ID: "c1", MType: models.Counter, Delta: intPtr(10)},
			},
			expectedError: false,
		},
		{
			name:           "nil storage initialized",
			initialStorage: nil,
			metrics: &[]models.Metrics{
				{ID: "c1", MType: models.Counter, Delta: intPtr(10)},
			},
			expectedStorage: map[string]models.Metrics{
				"c1": {ID: "c1", MType: models.Counter, Delta: intPtr(10)},
			},
			expectedError: false,
		},
		{
			name: "mixed existing and new counters",
			initialStorage: map[string]models.Metrics{
				"c1": {ID: "c1", MType: models.Counter, Delta: intPtr(100)},
			},
			metrics: &[]models.Metrics{
				{ID: "c1", MType: models.Counter, Delta: intPtr(50)},
				{ID: "c2", MType: models.Counter, Delta: intPtr(25)},
			},
			expectedStorage: map[string]models.Metrics{
				"c1": {ID: "c1", MType: models.Counter, Delta: intPtr(150)},
				"c2": {ID: "c2", MType: models.Counter, Delta: intPtr(25)},
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &memoryMetricsRepository{
				storage: &storage.MemStorage{
					Metrics: tt.initialStorage,
				},
				mutex: &sync.RWMutex{},
			}

			err := repo.AddAll(tt.metrics)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, len(tt.expectedStorage), len(repo.storage.Metrics))
			for id, expectedMetric := range tt.expectedStorage {
				actualMetric, exists := repo.storage.Metrics[id]
				assert.True(t, exists)
				assert.Equal(t, expectedMetric.ID, actualMetric.ID)
				assert.Equal(t, expectedMetric.MType, actualMetric.MType)
				assert.Equal(t, *expectedMetric.Delta, *actualMetric.Delta)
			}
		})
	}
}

func TestMetricsRepository_Reset(t *testing.T) {
	storage := storage.NewMemStorage()
	repo, err := NewMemoryMetricsRepository(storage, &sync.RWMutex{})

	assert.NoError(t, err)

	tests := []struct {
		name           string
		initialMetrics map[string]models.Metrics
		resetMetric    models.Metrics
		expectedMetric models.Metrics
	}{
		{
			name:           "reset existing metric",
			initialMetrics: map[string]models.Metrics{"m1": {ID: "m1", Value: floatPtr(10)}},
			resetMetric:    models.Metrics{ID: "m1", Value: floatPtr(5)},
			expectedMetric: models.Metrics{ID: "m1", Value: floatPtr(5)},
		},
		{
			name:           "reset non-existing metric",
			initialMetrics: map[string]models.Metrics{},
			resetMetric:    models.Metrics{ID: "m2", Value: floatPtr(7)},
			expectedMetric: models.Metrics{ID: "m2", Value: floatPtr(7)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage.Metrics = make(map[string]models.Metrics)
			for k, v := range tt.initialMetrics {
				storage.Metrics[k] = v
			}

			err := repo.Reset(tt.resetMetric)
			assert.NoError(t, err)

			result, _ := repo.Get(tt.resetMetric.ID)
			assert.NotNil(t, result)
			assert.Equal(t, *tt.expectedMetric.Value, *result.Value)
		})
	}
}

func TestMemoryMetricsRepository_ResetAll(t *testing.T) {
	tests := []struct {
		name            string
		initialStorage  map[string]models.Metrics
		metrics         *[]models.Metrics
		expectedStorage map[string]models.Metrics
		expectedError   bool
	}{
		{
			name:           "add new gauges",
			initialStorage: map[string]models.Metrics{},
			metrics: &[]models.Metrics{
				{ID: "g1", MType: models.Gauge, Value: floatPtr(3.14)},
				{ID: "g2", MType: models.Gauge, Value: floatPtr(2.71)},
			},
			expectedStorage: map[string]models.Metrics{
				"g1": {ID: "g1", MType: models.Gauge, Value: floatPtr(3.14)},
				"g2": {ID: "g2", MType: models.Gauge, Value: floatPtr(2.71)},
			},
			expectedError: false,
		},
		{
			name: "update existing gauges",
			initialStorage: map[string]models.Metrics{
				"g1": {ID: "g1", MType: models.Gauge, Value: floatPtr(1.0)},
				"g2": {ID: "g2", MType: models.Gauge, Value: floatPtr(2.0)},
			},
			metrics: &[]models.Metrics{
				{ID: "g1", MType: models.Gauge, Value: floatPtr(10.5)},
				{ID: "g2", MType: models.Gauge, Value: floatPtr(20.7)},
				{ID: "g3", MType: models.Gauge, Value: floatPtr(30.1)},
			},
			expectedStorage: map[string]models.Metrics{
				"g1": {ID: "g1", MType: models.Gauge, Value: floatPtr(10.5)},
				"g2": {ID: "g2", MType: models.Gauge, Value: floatPtr(20.7)},
				"g3": {ID: "g3", MType: models.Gauge, Value: floatPtr(30.1)},
			},
			expectedError: false,
		},
		{
			name: "empty metrics",
			initialStorage: map[string]models.Metrics{
				"g1": {ID: "g1", MType: models.Gauge, Value: floatPtr(3.14)},
			},
			metrics: &[]models.Metrics{},
			expectedStorage: map[string]models.Metrics{
				"g1": {ID: "g1", MType: models.Gauge, Value: floatPtr(3.14)},
			},
			expectedError: false,
		},
		{
			name:           "nil storage initialized",
			initialStorage: nil,
			metrics: &[]models.Metrics{
				{ID: "g1", MType: models.Gauge, Value: floatPtr(3.14)},
			},
			expectedStorage: map[string]models.Metrics{
				"g1": {ID: "g1", MType: models.Gauge, Value: floatPtr(3.14)},
			},
			expectedError: false,
		},
		{
			name: "mixed existing and new gauges",
			initialStorage: map[string]models.Metrics{
				"g1": {ID: "g1", MType: models.Gauge, Value: floatPtr(100.0)},
			},
			metrics: &[]models.Metrics{
				{ID: "g1", MType: models.Gauge, Value: floatPtr(50.5)},
				{ID: "g2", MType: models.Gauge, Value: floatPtr(25.3)},
			},
			expectedStorage: map[string]models.Metrics{
				"g1": {ID: "g1", MType: models.Gauge, Value: floatPtr(50.5)},
				"g2": {ID: "g2", MType: models.Gauge, Value: floatPtr(25.3)},
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &memoryMetricsRepository{
				storage: &storage.MemStorage{
					Metrics: tt.initialStorage,
				},
				mutex: &sync.RWMutex{},
			}

			err := repo.ResetAll(tt.metrics)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, len(tt.expectedStorage), len(repo.storage.Metrics))
			for id, expectedMetric := range tt.expectedStorage {
				actualMetric, exists := repo.storage.Metrics[id]
				assert.True(t, exists)
				assert.Equal(t, expectedMetric.ID, actualMetric.ID)
				assert.Equal(t, expectedMetric.MType, actualMetric.MType)
				assert.Equal(t, *expectedMetric.Value, *actualMetric.Value)
			}
		})
	}
}

func TestMetricsRepository_GetAll(t *testing.T) {
	tests := []struct {
		name           string
		initialMetrics map[string]models.Metrics
		expected       map[string]any
	}{
		{
			name:           "empty storage",
			initialMetrics: map[string]models.Metrics{},
			expected:       map[string]any{},
		},
		{
			name: "single counter metric",
			initialMetrics: map[string]models.Metrics{
				"c1": {
					ID:    "c1",
					MType: string(metric.CounterType),
					Delta: intPtr(5),
				},
			},
			expected: map[string]any{
				"c1": int64(5),
			},
		},
		{
			name: "single gauge metric",
			initialMetrics: map[string]models.Metrics{
				"g1": {
					ID:    "g1",
					MType: string(metric.GaugeType),
					Value: floatPtr(42.5),
				},
			},
			expected: map[string]any{
				"g1": float64(42.5),
			},
		},
		{
			name: "mixed metrics",
			initialMetrics: map[string]models.Metrics{
				"c1": {
					ID:    "c1",
					MType: string(metric.CounterType),
					Delta: intPtr(3),
				},
				"g1": {
					ID:    "g1",
					MType: string(metric.GaugeType),
					Value: floatPtr(99.9),
				},
			},
			expected: map[string]any{
				"c1": int64(3),
				"g1": float64(99.9),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := storage.NewMemStorage()
			st.Metrics = tt.initialMetrics
			repo := &memoryMetricsRepository{storage: st}

			result, _ := repo.GetAll()

			assert.NotNil(t, result)
			assert.Equal(t, tt.expected, *result)
		})
	}
}

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

func floatPtr(v float64) *float64 {
	return &v
}

func intPtr(value int64) *int64 {
	return &value
}
