package handler

import (
	"github.com/gabkaclassic/metrics/internal/model"
	"github.com/gabkaclassic/metrics/internal/service"
	"net/http"
	"net/http/httptest"

	api_error "github.com/gabkaclassic/metrics/internal/error"
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
)

type MockMetricsService struct {
	SaveFunc   func(id, metricType, rawValue string) *api_error.ApiError
	GetFunc    func(metricID string, metricType string) (any, *api_error.ApiError)
	GetAllFunc func() *map[string]any
}

func (m *MockMetricsService) Save(id, metricType, rawValue string) *api_error.ApiError {
	if m.SaveFunc != nil {
		return m.SaveFunc(id, metricType, rawValue)
	}
	return nil
}

func (m *MockMetricsService) Get(metricID string, metricType string) (any, *api_error.ApiError) {
	if m.GetFunc != nil {
		return m.GetFunc(metricID, metricType)
	}
	return nil, nil
}

func (m *MockMetricsService) GetAll() *map[string]any {
	if m.GetAllFunc != nil {
		return m.GetAllFunc()
	}
	return nil
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
		mockGet        func(metricID, metricType string) (any, *api_error.ApiError)
		expectStatus   int
		expectBody     *string
		expectErrorMsg string
	}{
		{
			name:   "valid GET gauge",
			method: http.MethodGet,
			pathVals: map[string]string{
				"id":   "m1",
				"type": models.Gauge,
			},
			mockGet: func(metricID, metricType string) (any, *api_error.ApiError) {
				assert.Equal(t, "m1", metricID)
				assert.Equal(t, models.Gauge, metricType)
				val := 42.0
				return &val, nil
			},
			expectStatus: http.StatusOK,
			expectBody:   strPtr("42\n"),
		},
		{
			name:   "valid GET counter",
			method: http.MethodGet,
			pathVals: map[string]string{
				"id":   "c1",
				"type": models.Counter,
			},
			mockGet: func(metricID, metricType string) (any, *api_error.ApiError) {
				assert.Equal(t, "c1", metricID)
				assert.Equal(t, models.Counter, metricType)
				val := int64(7)
				return &val, nil
			},
			expectStatus: http.StatusOK,
			expectBody:   strPtr("7\n"),
		},
		{
			name:   "service returns error",
			method: http.MethodGet,
			pathVals: map[string]string{
				"id":   "m2",
				"type": models.Gauge,
			},
			mockGet: func(metricID, metricType string) (any, *api_error.ApiError) {
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
				assert.Equal(t, *tt.expectBody, rr.Body.String())
			}
			if tt.expectErrorMsg != "" {
				assert.Contains(t, rr.Body.String(), tt.expectErrorMsg)
			}
		})
	}
}

func TestMetricsHandler_GetAll(t *testing.T) {
	tests := []struct {
		name           string
		mockReturn     *map[string]any
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "nil metrics",
			mockReturn:     nil,
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Metrics not found",
		},
		{
			name:           "empty metrics",
			mockReturn:     &map[string]any{},
			expectedStatus: http.StatusOK,
			expectedBody:   "<h1>Metrics</h1>",
		},
		{
			name: "metrics with values",
			mockReturn: &map[string]any{
				"c1": int64(10),
				"g1": float64(3.14),
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "<td>c1</td><td>10</td>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockMetricsService{
				GetAllFunc: func() *map[string]any {
					return tt.mockReturn
				},
			}
			handler := NewMetricsHandler(mockService)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()

			handler.GetAll(rec, req)

			res := rec.Result()
			defer res.Body.Close()

			body, _ := io.ReadAll(res.Body)

			assert.Equal(t, tt.expectedStatus, res.StatusCode)
			assert.Contains(t, string(body), tt.expectedBody)
		})
	}
}

func floatPtr(value float64) *float64 {
	return &value
}

func strPtr(value string) *string {
	return &value
}
