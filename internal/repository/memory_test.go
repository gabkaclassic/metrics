package repository

import (
	"errors"
	"sync"
	"testing"

	models "github.com/gabkaclassic/metrics/internal/model"
	"github.com/gabkaclassic/metrics/internal/storage"
	"github.com/gabkaclassic/metrics/pkg/metric"
	"github.com/stretchr/testify/assert"
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
			result, err := repo.Get(t.Context(), tt.metricID)

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
		metric      *models.Metrics
		updateFunc  func(metric models.Metrics) error
		expectError bool
	}{
		{
			name:   "successful update",
			metric: &models.Metrics{ID: "m1", Value: floatPtr(10)},
			updateFunc: func(metric models.Metrics) error {
				repo.storage.Metrics[metric.ID] = metric
				return nil
			},
			expectError: false,
		},
		{
			name:   "update returns error",
			metric: &models.Metrics{ID: "m2", Value: floatPtr(20)},
			updateFunc: func(metric models.Metrics) error {
				return errors.New("update failed")
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.updateMetric(t.Context(), *tt.metric, tt.updateFunc)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				stored, _ := repo.Get(t.Context(), tt.metric.ID)
				assert.Equal(t, tt.metric, stored)
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
		metrics         []models.Metrics
		updateFn        func(metric []models.Metrics) error
		expectedError   bool
		expectedStorage map[string]models.Metrics
	}{
		{
			name:           "successful update",
			initialStorage: map[string]models.Metrics{},
			metrics: []models.Metrics{
				{ID: "test1", MType: models.Gauge, Value: floatPtr(1.0)},
			},
			updateFn: func(metrics []models.Metrics) error {
				return nil
			},
			expectedError:   false,
			expectedStorage: map[string]models.Metrics{},
		},
		{
			name:           "update function error",
			initialStorage: map[string]models.Metrics{},
			metrics: []models.Metrics{
				{ID: "test1", MType: models.Gauge, Value: floatPtr(1.0)},
			},
			updateFn: func(metrics []models.Metrics) error {
				return errors.New("update error")
			},
			expectedError:   true,
			expectedStorage: map[string]models.Metrics{},
		},
		{
			name:           "initialize nil storage",
			initialStorage: nil,
			metrics: []models.Metrics{
				{ID: "test1", MType: models.Gauge, Value: floatPtr(1.0)},
			},
			updateFn: func(metrics []models.Metrics) error {
				return nil
			},
			expectedError:   false,
			expectedStorage: map[string]models.Metrics{},
		},
		{
			name:           "empty metrics",
			initialStorage: map[string]models.Metrics{},
			metrics:        []models.Metrics{},
			updateFn: func(metrics []models.Metrics) error {
				return nil
			},
			expectedError:   false,
			expectedStorage: map[string]models.Metrics{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo.storage.Metrics = tt.initialStorage

			err := repo.updateMetrics(t.Context(), tt.metrics, tt.updateFn)

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
				err := repo.Add(t.Context(), m)
				assert.NoError(t, err)
			}

			err = repo.Add(t.Context(), tt.addMetric)
			assert.NoError(t, err)

			result, err := repo.Get(t.Context(), tt.addMetric.ID)
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
		metrics         []models.Metrics
		expectedStorage map[string]models.Metrics
		expectedError   bool
	}{
		{
			name:           "add new counters",
			initialStorage: map[string]models.Metrics{},
			metrics: []models.Metrics{
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
			metrics: []models.Metrics{
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
			metrics: []models.Metrics{},
			expectedStorage: map[string]models.Metrics{
				"c1": {ID: "c1", MType: models.Counter, Delta: intPtr(10)},
			},
			expectedError: false,
		},
		{
			name:           "nil storage initialized",
			initialStorage: nil,
			metrics: []models.Metrics{
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
			metrics: []models.Metrics{
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

			err := repo.AddAll(t.Context(), tt.metrics)

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

			err := repo.Reset(t.Context(), tt.resetMetric)
			assert.NoError(t, err)

			result, _ := repo.Get(t.Context(), tt.resetMetric.ID)
			assert.NotNil(t, result)
			assert.Equal(t, *tt.expectedMetric.Value, *result.Value)
		})
	}
}

func TestMemoryMetricsRepository_ResetAll(t *testing.T) {
	tests := []struct {
		name            string
		initialStorage  map[string]models.Metrics
		metrics         []models.Metrics
		expectedStorage map[string]models.Metrics
		expectedError   bool
	}{
		{
			name:           "add new gauges",
			initialStorage: map[string]models.Metrics{},
			metrics: []models.Metrics{
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
			metrics: []models.Metrics{
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
			metrics: []models.Metrics{},
			expectedStorage: map[string]models.Metrics{
				"g1": {ID: "g1", MType: models.Gauge, Value: floatPtr(3.14)},
			},
			expectedError: false,
		},
		{
			name:           "nil storage initialized",
			initialStorage: nil,
			metrics: []models.Metrics{
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
			metrics: []models.Metrics{
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

			err := repo.ResetAll(t.Context(), tt.metrics)

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

			result, _ := repo.GetAll(t.Context())

			assert.NotNil(t, result)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMemoryMetricsRepository_GetAllMetrics(t *testing.T) {
	tests := []struct {
		name           string
		initialStorage map[string]models.Metrics
		expectedCount  int
		expectedError  bool
	}{
		{
			name:           "empty storage",
			initialStorage: map[string]models.Metrics{},
			expectedCount:  0,
			expectedError:  false,
		},
		{
			name: "storage with counters",
			initialStorage: map[string]models.Metrics{
				"c1": {ID: "c1", MType: models.Counter, Delta: intPtr(10)},
				"c2": {ID: "c2", MType: models.Counter, Delta: intPtr(5)},
			},
			expectedCount: 2,
			expectedError: false,
		},
		{
			name: "storage with gauges",
			initialStorage: map[string]models.Metrics{
				"g1": {ID: "g1", MType: models.Gauge, Value: floatPtr(3.14)},
				"g2": {ID: "g2", MType: models.Gauge, Value: floatPtr(2.71)},
			},
			expectedCount: 2,
			expectedError: false,
		},
		{
			name: "storage with mixed metrics",
			initialStorage: map[string]models.Metrics{
				"c1": {ID: "c1", MType: models.Counter, Delta: intPtr(10)},
				"g1": {ID: "g1", MType: models.Gauge, Value: floatPtr(3.14)},
				"c2": {ID: "c2", MType: models.Counter, Delta: intPtr(5)},
				"g2": {ID: "g2", MType: models.Gauge, Value: floatPtr(2.71)},
			},
			expectedCount: 4,
			expectedError: false,
		},
		{
			name:           "nil storage",
			initialStorage: nil,
			expectedCount:  0,
			expectedError:  false,
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

			result, err := repo.GetAllMetrics(t.Context())

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedCount, len(result))

				if tt.initialStorage != nil {
					for id, expectedMetric := range tt.initialStorage {
						found := false
						for _, actualMetric := range result {
							if actualMetric.ID == id {
								found = true
								assert.Equal(t, expectedMetric.ID, actualMetric.ID)
								assert.Equal(t, expectedMetric.MType, actualMetric.MType)
								if expectedMetric.Delta != nil {
									assert.Equal(t, *expectedMetric.Delta, *actualMetric.Delta)
								}
								if expectedMetric.Value != nil {
									assert.Equal(t, *expectedMetric.Value, *actualMetric.Value)
								}
								break
							}
						}
						assert.True(t, found, "Metric %s not found in result", id)
					}
				}
			}
		})
	}
}

func floatPtr(v float64) *float64 {
	return &v
}

func intPtr(value int64) *int64 {
	return &value
}
