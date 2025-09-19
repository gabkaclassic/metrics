package handler

import (
	"encoding/json"
	"github.com/gabkaclassic/metrics/internal/model"
	"github.com/gabkaclassic/metrics/internal/service"
	"net/http"
	"net/http/httptest"

	api_error "github.com/gabkaclassic/metrics/internal/error"
	"github.com/stretchr/testify/assert"
	"testing"
)

type MockMetricsService struct {
	SaveFunc func(id, metricType, rawValue string) *api_error.ApiError
	GetFunc  func(metricID string) (*models.Metrics, *api_error.ApiError)
}

func (m *MockMetricsService) Save(id, metricType, rawValue string) *api_error.ApiError {
	if m.SaveFunc != nil {
		return m.SaveFunc(id, metricType, rawValue)
	}
	return nil
}

func (m *MockMetricsService) Get(metricID string) (*models.Metrics, *api_error.ApiError) {
	if m.GetFunc != nil {
		return m.GetFunc(metricID)
	}
	return nil, nil
}
func TestNewMetricsHandler(t *testing.T) {
	var validService service.MetricsService = &MockMetricsService{}

	tests := []struct {
		name        string
		service     service.MetricsService
		expectPanic bool
	}{
		{
			name:        "valid service",
			service:     validService,
			expectPanic: false,
		},
		{
			name:        "nil service",
			service:     nil,
			expectPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectPanic {
				assert.Panics(t, func() {
					NewMetricsHandler(tt.service)
				})
			} else {
				handler := NewMetricsHandler(tt.service)
				assert.NotNil(t, handler)
				assert.Equal(t, tt.service, handler.service)
			}
		})
	}
}

func TestMetricsHandler_Save(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		pathVals       map[string]string
		mockSave       func(id, metricType, rawValue string) *api_error.ApiError
		expectStatus   int
		expectErrorMsg string
	}{
		{
			name:   "valid POST",
			method: http.MethodPost,
			pathVals: map[string]string{
				"id":    "m1",
				"type":  models.Counter,
				"value": "10",
			},
			mockSave: func(id, metricType, rawValue string) *api_error.ApiError {
				assert.Equal(t, "m1", id)
				assert.Equal(t, models.Counter, metricType)
				assert.Equal(t, "10", rawValue)
				return nil
			},
			expectStatus: http.StatusOK,
		},
		{
			name:         "invalid method",
			method:       http.MethodGet,
			pathVals:     map[string]string{},
			mockSave:     nil,
			expectStatus: http.StatusMethodNotAllowed,
		},
		{
			name:   "service returns error",
			method: http.MethodPost,
			pathVals: map[string]string{
				"id":    "m2",
				"type":  models.Gauge,
				"value": "abc",
			},
			mockSave: func(id, metricType, rawValue string) *api_error.ApiError {
				return api_error.BadRequest("invalid metric value")
			},
			expectStatus:   http.StatusBadRequest,
			expectErrorMsg: "invalid metric value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockMetricsService{
				SaveFunc: tt.mockSave,
			}
			handler := NewMetricsHandler(mockService)

			req := httptest.NewRequest(tt.method, "/", nil)
			for k, v := range tt.pathVals {
				req.SetPathValue(k, v)
			}
			rr := httptest.NewRecorder()

			handler.Save(rr, req)

			assert.Equal(t, tt.expectStatus, rr.Code)
			if tt.expectErrorMsg != "" {
				assert.Contains(t, rr.Body.String(), tt.expectErrorMsg)
			}
		})
	}
}

func TestMetricsHandler_Get(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		pathVals       map[string]string
		mockGet        func(metricID string) (*models.Metrics, *api_error.ApiError)
		expectStatus   int
		expectBody     *models.Metrics
		expectErrorMsg string
	}{
		{
			name:   "valid GET",
			method: http.MethodGet,
			pathVals: map[string]string{
				"id": "m1",
			},
			mockGet: func(metricID string) (*models.Metrics, *api_error.ApiError) {
				assert.Equal(t, "m1", metricID)
				return &models.Metrics{ID: "m1", Value: floatPtr(42)}, nil
			},
			expectStatus: http.StatusOK,
			expectBody:   &models.Metrics{ID: "m1", Value: floatPtr(42)},
		},
		{
			name:         "invalid method",
			method:       http.MethodPost,
			pathVals:     map[string]string{},
			mockGet:      nil,
			expectStatus: http.StatusMethodNotAllowed,
		},
		{
			name:   "service returns error",
			method: http.MethodGet,
			pathVals: map[string]string{
				"id": "m2",
			},
			mockGet: func(metricID string) (*models.Metrics, *api_error.ApiError) {
				return nil, api_error.NotFound("metric m2 not found")
			},
			expectStatus:   http.StatusNotFound,
			expectErrorMsg: "metric m2 not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockMetricsService{
				GetFunc: tt.mockGet,
			}
			handler := NewMetricsHandler(mockService)

			req := httptest.NewRequest(tt.method, "/", nil)
			for k, v := range tt.pathVals {
				req.SetPathValue(k, v)
			}
			rr := httptest.NewRecorder()

			handler.Get(rr, req)

			assert.Equal(t, tt.expectStatus, rr.Code)
			if tt.expectBody != nil {
				var body models.Metrics
				err := json.NewDecoder(rr.Body).Decode(&body)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectBody, &body)
			}
			if tt.expectErrorMsg != "" {
				assert.Contains(t, rr.Body.String(), tt.expectErrorMsg)
			}
		})
	}
}

func floatPtr(value float64) *float64 {
	return &value
}
