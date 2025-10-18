package repository

import (
	"errors"
	"sync"
	"testing"

	"github.com/gabkaclassic/metrics/internal/model"
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

func floatPtr(v float64) *float64 {
	return &v
}

func intPtr(value int64) *int64 {
	return &value
}
