package repository

import (
	"errors"
	"github.com/gabkaclassic/metrics/internal/model"
	"github.com/gabkaclassic/metrics/internal/storage"
	"github.com/gabkaclassic/metrics/pkg/metric"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewMetricsRepository(t *testing.T) {
	tests := []struct {
		name        string
		storage     *storage.MemStorage
		expectPanic bool
	}{
		{
			name:        "valid storage",
			storage:     storage.NewMemStorage(),
			expectPanic: false,
		},
		{
			name:        "nil storage",
			storage:     nil,
			expectPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectPanic {
				assert.Panics(t, func() {
					NewMetricsRepository(tt.storage)
				})
			} else {
				repo := NewMetricsRepository(tt.storage)
				assert.NotNil(t, repo)
			}
		})
	}
}

func TestMetricsRepository_Get(t *testing.T) {
	storage := storage.NewMemStorage()
	storage.Metrics["existing"] = models.Metrics{ID: "existing", Value: floatPtr(42)}
	repo := &metricsRepository{storage: storage}

	tests := []struct {
		name        string
		metricID    string
		expectValue *models.Metrics
		expectNil   bool
	}{
		{
			name:        "metric exists",
			metricID:    "existing",
			expectValue: &models.Metrics{ID: "existing", Value: floatPtr(42)},
			expectNil:   false,
		},
		{
			name:        "metric does not exist",
			metricID:    "missing",
			expectValue: nil,
			expectNil:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _ := repo.Get(tt.metricID)
			if tt.expectNil {
				assert.Nil(t, result)
			} else {
				assert.Equal(t, tt.expectValue, result)
			}
		})
	}
}

func TestMetricsRepository_updateMetric(t *testing.T) {
	storage := storage.NewMemStorage()
	repo := &metricsRepository{storage: storage}

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
	storage := storage.NewMemStorage()
	repo := &metricsRepository{storage: storage}

	tests := []struct {
		name           string
		initialMetrics map[string]models.Metrics
		addMetric      models.Metrics
		expectedMetric models.Metrics
	}{
		{
			name:           "add new metric",
			initialMetrics: map[string]models.Metrics{},
			addMetric:      models.Metrics{ID: "m1", Delta: intPtr(5)},
			expectedMetric: models.Metrics{ID: "m1", Delta: intPtr(5)},
		},
		{
			name: "update existing metric",
			initialMetrics: map[string]models.Metrics{
				"m2": {ID: "m2", Delta: intPtr(3)},
			},
			addMetric:      models.Metrics{ID: "m2", Delta: intPtr(2)},
			expectedMetric: models.Metrics{ID: "m2", Delta: intPtr(5)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage.Mutex.Lock()
			storage.Metrics = make(map[string]models.Metrics)
			for k, v := range tt.initialMetrics {
				storage.Metrics[k] = v
			}
			storage.Mutex.Unlock()

			err := repo.Add(tt.addMetric)
			assert.NoError(t, err)

			result, _ := repo.Get(tt.addMetric.ID)
			assert.NotNil(t, result)
			assert.Equal(t, *tt.expectedMetric.Delta, *result.Delta)
		})
	}
}

func TestMetricsRepository_Reset(t *testing.T) {
	storage := storage.NewMemStorage()
	repo := &metricsRepository{storage: storage}

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
			storage.Mutex.Lock()
			storage.Metrics = make(map[string]models.Metrics)
			for k, v := range tt.initialMetrics {
				storage.Metrics[k] = v
			}
			storage.Mutex.Unlock()

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
			repo := &metricsRepository{storage: st}

			result := repo.GetAll()

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
