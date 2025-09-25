package service

import (
	"errors"
	"testing"

	models "github.com/gabkaclassic/metrics/internal/model"
	"github.com/gabkaclassic/metrics/internal/repository"
	"github.com/stretchr/testify/assert"
)

func TestNewMetricsService(t *testing.T) {
	mockRepo := repository.NewMockMetricsRepository(t)

	tests := []struct {
		name        string
		repository  repository.MetricsRepository
		expectError bool
	}{
		{
			name:        "valid repository",
			repository:  mockRepo,
			expectError: false,
		},
		{
			name:        "nil repository",
			repository:  nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, err := NewMetricsService(tt.repository)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, svc)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, svc)
			}
		})
	}
}

func TestMetricsService_Get(t *testing.T) {
	tests := []struct {
		name           string
		metricID       string
		metricType     string
		setupMock      func(m *repository.MockMetricsRepository)
		expectValue    any
		expectAPIError bool
		expectNotFound bool
	}{
		{
			name:       "metric exists with correct type (gauge)",
			metricID:   "m1",
			metricType: models.Gauge,
			setupMock: func(m *repository.MockMetricsRepository) {
				m.EXPECT().Get("m1").
					Return(&models.Metrics{ID: "m1", MType: models.Gauge, Value: floatPtr(10)}, nil)
			},
			expectValue:    floatPtr(10),
			expectAPIError: false,
			expectNotFound: false,
		},
		{
			name:       "metric exists but wrong type",
			metricID:   "m1",
			metricType: models.Counter,
			setupMock: func(m *repository.MockMetricsRepository) {
				m.EXPECT().Get("m1").
					Return(&models.Metrics{ID: "m1", MType: models.Gauge, Value: floatPtr(10)}, nil)
			},
			expectValue:    nil,
			expectAPIError: true,
			expectNotFound: true,
		},
		{
			name:       "metric does not exist",
			metricID:   "m2",
			metricType: models.Gauge,
			setupMock: func(m *repository.MockMetricsRepository) {
				m.EXPECT().Get("m2").
					Return(nil, nil)
			},
			expectValue:    nil,
			expectAPIError: true,
			expectNotFound: true,
		},
		{
			name:       "repository returns error",
			metricID:   "m3",
			metricType: models.Gauge,
			setupMock: func(m *repository.MockMetricsRepository) {
				m.EXPECT().Get("m3").
					Return(nil, errors.New("db error"))
			},
			expectValue:    nil,
			expectAPIError: true,
			expectNotFound: false,
		},
		{
			name:       "metric exists with correct type (counter)",
			metricID:   "m4",
			metricType: models.Counter,
			setupMock: func(m *repository.MockMetricsRepository) {
				m.EXPECT().Get("m4").
					Return(&models.Metrics{ID: "m4", MType: models.Counter, Delta: intPtr(42)}, nil)
			},
			expectValue:    intPtr(42),
			expectAPIError: false,
			expectNotFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := repository.NewMockMetricsRepository(t)
			tt.setupMock(mockRepo)

			svc, err := NewMetricsService(mockRepo)
			assert.NoError(t, err)

			result, apiErr := svc.Get(tt.metricID, tt.metricType)

			if tt.expectAPIError {
				assert.NotNil(t, apiErr)
				if tt.expectNotFound {
					assert.Contains(t, apiErr.Message, "not found")
					assert.Contains(t, apiErr.Message, tt.metricID)
					assert.Contains(t, apiErr.Message, tt.metricType)
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
		setupMock     func(m *repository.MockMetricsRepository)
		expectError   bool
		errorContains string
	}{
		{
			name:       "valid counter",
			id:         "c1",
			metricType: models.Counter,
			rawValue:   "10",
			setupMock: func(m *repository.MockMetricsRepository) {
				m.EXPECT().
					Add(models.Metrics{
						ID:    "c1",
						MType: models.Counter,
						Delta: intPtr(10),
					}).
					Return(nil)
			},
			expectError: false,
		},
		{
			name:          "invalid counter",
			id:            "c2",
			metricType:    models.Counter,
			rawValue:      "abc",
			setupMock:     func(m *repository.MockMetricsRepository) {},
			expectError:   true,
			errorContains: "invalid metric value",
		},
		{
			name:       "valid gauge",
			id:         "g1",
			metricType: models.Gauge,
			rawValue:   "3.14",
			setupMock: func(m *repository.MockMetricsRepository) {
				m.EXPECT().
					Reset(models.Metrics{
						ID:    "g1",
						MType: models.Gauge,
						Value: floatPtr(3.14),
					}).
					Return(nil)
			},
			expectError: false,
		},
		{
			name:          "invalid gauge",
			id:            "g2",
			metricType:    models.Gauge,
			rawValue:      "xyz",
			setupMock:     func(m *repository.MockMetricsRepository) {},
			expectError:   true,
			errorContains: "invalid metric value",
		},
		{
			name:          "invalid type",
			id:            "x1",
			metricType:    "unknown",
			rawValue:      "123",
			setupMock:     func(m *repository.MockMetricsRepository) {},
			expectError:   true,
			errorContains: "invalid metric type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := repository.NewMockMetricsRepository(t)
			tt.setupMock(mockRepo)

			svc, err := NewMetricsService(mockRepo)
			assert.NoError(t, err)

			apiErr := svc.Save(tt.id, tt.metricType, tt.rawValue)

			if tt.expectError {
				assert.NotNil(t, apiErr)
				assert.Contains(t, apiErr.Message, tt.errorContains)
			} else {
				assert.Nil(t, apiErr)
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
			mockRepo := repository.NewMockMetricsRepository(t)
			mockRepo.EXPECT().
				GetAll().
				Return(tt.mockReturn)

			svc, err := NewMetricsService(mockRepo)
			assert.NoError(t, err)

			result := svc.GetAll()

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
