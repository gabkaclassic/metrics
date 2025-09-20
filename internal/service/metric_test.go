package service

import (
	"errors"
	"github.com/gabkaclassic/metrics/internal/model"
	"github.com/gabkaclassic/metrics/internal/repository"
	"github.com/stretchr/testify/assert"
	"testing"
)

type MockMetricsRepository struct {
	AddFunc    func(metric models.Metrics) error
	ResetFunc  func(metric models.Metrics) error
	GetFunc    func(metricID string) (*models.Metrics, error)
	GetAllFunc func() *map[string]any
}

func (m *MockMetricsRepository) Add(metric models.Metrics) error {
	if m.AddFunc != nil {
		return m.AddFunc(metric)
	}
	return nil
}

func (m *MockMetricsRepository) Reset(metric models.Metrics) error {
	if m.ResetFunc != nil {
		return m.ResetFunc(metric)
	}
	return nil
}

func (m *MockMetricsRepository) Get(metricID string) (*models.Metrics, error) {
	if m.GetFunc != nil {
		return m.GetFunc(metricID)
	}
	return nil, nil
}

func (m *MockMetricsRepository) GetAll() *map[string]any {
	if m.GetAllFunc != nil {
		return m.GetAllFunc()
	}
	return nil
}

func TestNewMetricsService(t *testing.T) {
	mockRepo := &MockMetricsRepository{}

	tests := []struct {
		name        string
		repository  repository.MetricsRepository
		expectPanic bool
	}{
		{
			name:        "valid repository",
			repository:  mockRepo,
			expectPanic: false,
		},
		{
			name:        "nil repository",
			repository:  nil,
			expectPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectPanic {
				assert.Panics(t, func() {
					NewMetricsService(tt.repository)
				})
			} else {
				service := NewMetricsService(tt.repository)
				assert.NotNil(t, service)
			}
		})
	}
}

func TestMetricsService_Get(t *testing.T) {
	tests := []struct {
		name           string
		metricID       string
		metricType     string
		mockGet        func(metricID string) (*models.Metrics, error)
		expectValue    any
		expectApiError bool
		expectNotFound bool
	}{
		{
			name:       "metric exists with correct type (gauge)",
			metricID:   "m1",
			metricType: models.Gauge,
			mockGet: func(metricID string) (*models.Metrics, error) {
				return &models.Metrics{ID: "m1", MType: models.Gauge, Value: floatPtr(10)}, nil
			},
			expectValue:    floatPtr(10),
			expectApiError: false,
			expectNotFound: false,
		},
		{
			name:       "metric exists but wrong type",
			metricID:   "m1",
			metricType: models.Counter,
			mockGet: func(metricID string) (*models.Metrics, error) {
				return &models.Metrics{ID: "m1", MType: models.Gauge, Value: floatPtr(10)}, nil
			},
			expectValue:    nil,
			expectApiError: true,
			expectNotFound: true,
		},
		{
			name:       "metric does not exist",
			metricID:   "m2",
			metricType: models.Gauge,
			mockGet: func(metricID string) (*models.Metrics, error) {
				return nil, nil
			},
			expectValue:    nil,
			expectApiError: true,
			expectNotFound: true,
		},
		{
			name:       "repository returns error",
			metricID:   "m3",
			metricType: models.Gauge,
			mockGet: func(metricID string) (*models.Metrics, error) {
				return nil, errors.New("db error")
			},
			expectValue:    nil,
			expectApiError: true,
			expectNotFound: false,
		},
		{
			name:       "metric exists with correct type (counter)",
			metricID:   "m4",
			metricType: models.Counter,
			mockGet: func(metricID string) (*models.Metrics, error) {
				return &models.Metrics{ID: "m4", MType: models.Counter, Delta: intPtr(42)}, nil
			},
			expectValue:    intPtr(42),
			expectApiError: false,
			expectNotFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockMetricsRepository{
				GetFunc: tt.mockGet,
			}
			service := NewMetricsService(mockRepo)

			result, apiErr := service.Get(tt.metricID, tt.metricType)

			if tt.expectApiError {
				assert.NotNil(t, apiErr)
				if tt.expectNotFound {
					assert.Contains(t, apiErr.Message, "not found")
					assert.Contains(t, apiErr.Message, tt.metricID)
					assert.Contains(t, apiErr.Message, tt.metricType)
				} else {
					assert.Contains(t, apiErr.Message, "not found")
				}
			} else {
				assert.Nil(t, apiErr)
				assert.Equal(t, tt.expectValue, result)
			}
		})
	}
}

func intPtr(value int64) *int64 {
	return &value
}

func TestMetricsService_Save(t *testing.T) {
	tests := []struct {
		name          string
		id            string
		metricType    string
		rawValue      string
		mockAdd       func(metric models.Metrics) error
		mockReset     func(metric models.Metrics) error
		expectError   bool
		errorContains string
	}{
		{
			name:       "valid counter",
			id:         "c1",
			metricType: models.Counter,
			rawValue:   "10",
			mockAdd: func(metric models.Metrics) error {
				assert.Equal(t, int64(10), *metric.Delta)
				assert.Equal(t, "c1", metric.ID)
				assert.Equal(t, models.Counter, metric.MType)
				return nil
			},
			expectError: false,
		},
		{
			name:          "invalid counter",
			id:            "c2",
			metricType:    models.Counter,
			rawValue:      "abc",
			expectError:   true,
			errorContains: "invalid metric value",
		},
		{
			name:       "valid gauge",
			id:         "g1",
			metricType: models.Gauge,
			rawValue:   "3.14",
			mockReset: func(metric models.Metrics) error {
				assert.Equal(t, 3.14, *metric.Value)
				assert.Equal(t, "g1", metric.ID)
				assert.Equal(t, models.Gauge, metric.MType)
				return nil
			},
			expectError: false,
		},
		{
			name:          "invalid gauge",
			id:            "g2",
			metricType:    models.Gauge,
			rawValue:      "xyz",
			expectError:   true,
			errorContains: "invalid metric value",
		},
		{
			name:          "invalid type",
			id:            "x1",
			metricType:    "unknown",
			rawValue:      "123",
			expectError:   true,
			errorContains: "invalid metric type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockMetricsRepository{
				AddFunc:   tt.mockAdd,
				ResetFunc: tt.mockReset,
			}
			service := NewMetricsService(mockRepo)

			err := service.Save(tt.id, tt.metricType, tt.rawValue)

			if tt.expectError {
				assert.NotNil(t, err)
				assert.Contains(t, err.Message, tt.errorContains)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestMetricsService_GetAll(t *testing.T) {
	tests := []struct {
		name       string
		mockReturn *map[string]any
		expected   map[string]any
	}{
		{
			name:       "empty repository",
			mockReturn: &map[string]any{},
			expected:   map[string]any{},
		},
		{
			name: "repository with metrics",
			mockReturn: &map[string]any{
				"c1": int64(10),
				"g1": float64(3.14),
			},
			expected: map[string]any{
				"c1": int64(10),
				"g1": float64(3.14),
			},
		},
		{
			name:       "repository returns nil",
			mockReturn: nil,
			expected:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockMetricsRepository{
				GetAllFunc: func() *map[string]any {
					return tt.mockReturn
				},
			}

			service := NewMetricsService(mockRepo)

			result := service.GetAll()

			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, tt.expected, *result)
			}
		})
	}
}

func floatPtr(value float64) *float64 {
	return &value
}
