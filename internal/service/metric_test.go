package service

import (
	"errors"
	"net/http"
	"testing"

	models "github.com/gabkaclassic/metrics/internal/model"
	"github.com/gabkaclassic/metrics/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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

func TestMetricsService_GetStruct(t *testing.T) {
	tests := []struct {
		name           string
		metricID       string
		metricType     string
		mockGet        func(id string) (*models.Metrics, error)
		expectResult   *models.Metrics
		expectErrorMsg string
		expectStatus   int
	}{
		{
			name:       "metric found and type matches",
			metricID:   "m1",
			metricType: "counter",
			mockGet: func(id string) (*models.Metrics, error) {
				delta := int64(10)
				return &models.Metrics{ID: "m1", MType: "counter", Delta: &delta}, nil
			},
			expectResult: &models.Metrics{
				ID:    "m1",
				MType: "counter",
				Delta: func() *int64 { v := int64(10); return &v }(),
			},
			expectStatus: http.StatusOK,
		},
		{
			name:       "metric not found",
			metricID:   "m2",
			metricType: "gauge",
			mockGet: func(id string) (*models.Metrics, error) {
				return nil, nil
			},
			expectErrorMsg: "metric m2 gauge not found",
			expectStatus:   http.StatusNotFound,
		},
		{
			name:       "metric type mismatch",
			metricID:   "m3",
			metricType: "counter",
			mockGet: func(id string) (*models.Metrics, error) {
				val := 3.14
				return &models.Metrics{ID: "m3", MType: "gauge", Value: &val}, nil
			},
			expectErrorMsg: "metric m3 counter not found",
			expectStatus:   http.StatusNotFound,
		},
		{
			name:       "repository returns error (metric is nil)",
			metricID:   "m4",
			metricType: "counter",
			mockGet: func(id string) (*models.Metrics, error) {
				return nil, errors.New("db error")
			},
			expectErrorMsg: "metric m4 counter not found",
			expectStatus:   http.StatusNotFound,
		},
		{
			name:       "repository returns error but metric not nil",
			metricID:   "m5",
			metricType: "counter",
			mockGet: func(id string) (*models.Metrics, error) {
				delta := int64(1)
				return &models.Metrics{ID: "m5", MType: "counter", Delta: &delta}, errors.New("db error")
			},
			expectErrorMsg: "Get metric error",
			expectStatus:   http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := repository.NewMockMetricsRepository(t)
			mockRepo.EXPECT().
				Get(tt.metricID).
				RunAndReturn(tt.mockGet)

			svc, _ := NewMetricsService(mockRepo)

			result, apiErr := svc.GetStruct(tt.metricID, tt.metricType)

			if tt.expectStatus == http.StatusOK {
				require.Nil(t, apiErr)
				require.NotNil(t, result)
				assert.Equal(t, tt.expectResult.ID, result.ID)
				assert.Equal(t, tt.expectResult.MType, result.MType)
				assert.Equal(t, tt.expectResult.Delta, result.Delta)
				assert.Equal(t, tt.expectResult.Value, result.Value)
			} else {
				require.NotNil(t, apiErr)
				assert.Equal(t, tt.expectStatus, apiErr.Code)
				assert.Contains(t, apiErr.Message, tt.expectErrorMsg)
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

func TestMetricsService_SaveStruct(t *testing.T) {
	tests := []struct {
		name           string
		input          models.Metrics
		mockAdd        func(metric models.Metrics)
		mockReset      func(metric models.Metrics)
		expectErrorMsg string
		expectStatus   int
	}{
		{
			name: "counter metric calls Add",
			input: models.Metrics{
				ID:    "m1",
				MType: models.Counter,
				Delta: func() *int64 { v := int64(10); return &v }(),
			},
			mockAdd: func(metric models.Metrics) {
				assert.Equal(t, "m1", metric.ID)
				assert.Equal(t, models.Counter, metric.MType)
				assert.Equal(t, int64(10), *metric.Delta)
			},
			expectStatus: http.StatusOK,
		},
		{
			name: "gauge metric calls Reset",
			input: models.Metrics{
				ID:    "m2",
				MType: models.Gauge,
				Value: func() *float64 { v := 3.14; return &v }(),
			},
			mockReset: func(metric models.Metrics) {
				assert.Equal(t, "m2", metric.ID)
				assert.Equal(t, models.Gauge, metric.MType)
				assert.Equal(t, 3.14, *metric.Value)
			},
			expectStatus: http.StatusOK,
		},
		{
			name: "invalid metric type",
			input: models.Metrics{
				ID:    "m3",
				MType: "unknown",
			},
			expectErrorMsg: "invalid metric type: unknown",
			expectStatus:   http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := repository.NewMockMetricsRepository(t)

			if tt.input.MType == models.Counter {
				mockRepo.EXPECT().
					Add(mock.AnythingOfType("models.Metrics")).
					RunAndReturn(func(metric models.Metrics) error {
						if tt.mockAdd != nil {
							tt.mockAdd(metric)
						}
						return nil
					})
			}
			if tt.input.MType == models.Gauge {
				mockRepo.EXPECT().
					Reset(mock.AnythingOfType("models.Metrics")).
					RunAndReturn(func(metric models.Metrics) error {
						if tt.mockReset != nil {
							tt.mockReset(metric)
						}
						return nil
					})
			}

			svc, _ := NewMetricsService(mockRepo)

			apiErr := svc.SaveStruct(tt.input)

			if tt.expectStatus == http.StatusOK {
				assert.Nil(t, apiErr)
			} else {
				require.NotNil(t, apiErr)
				assert.Equal(t, tt.expectStatus, apiErr.Code)
				assert.Contains(t, apiErr.Message, tt.expectErrorMsg)
			}
		})
	}
}

func TestMetricsService_GetAll(t *testing.T) {
	tests := []struct {
		name          string
		mockReturn    *map[string]any
		expected      map[string]any
		expectedError error
	}{
		{
			name:          "empty repository",
			mockReturn:    &map[string]any{},
			expected:      map[string]any{},
			expectedError: nil,
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
			expectedError: nil,
		},
		{
			name:          "repository returns nil",
			mockReturn:    nil,
			expected:      nil,
			expectedError: nil,
		},
		{
			name:          "repository returns error",
			mockReturn:    nil,
			expected:      nil,
			expectedError: errors.New("some error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := repository.NewMockMetricsRepository(t)
			mockRepo.EXPECT().
				GetAll().
				Return(tt.mockReturn, tt.expectedError)

			svc, err := NewMetricsService(mockRepo)
			assert.NoError(t, err)

			result, err := svc.GetAll()

			if tt.expectedError == nil {
				assert.Nil(t, err)
			} else {
				assert.Error(t, err)
			}

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
