package handler

import (
	"net/http"
	"net/http/httptest"

	models "github.com/gabkaclassic/metrics/internal/model"
	"github.com/gabkaclassic/metrics/internal/service"

	"fmt"
	"io"
	"testing"

	api "github.com/gabkaclassic/metrics/internal/error"
	"github.com/stretchr/testify/assert"
)

func TestNewMetricsHandler(t *testing.T) {
	validService := service.NewMockMetricsService(t)

	tests := []struct {
		name        string
		service     service.MetricsService
		expectError bool
	}{
		{
			name:        "valid service",
			service:     validService,
			expectError: false,
		},
		{
			name:        "nil service",
			service:     nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, err := NewMetricsHandler(tt.service)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, handler)
			} else {
				assert.NoError(t, err)
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
		mockSave       func(id, metricType, rawValue string) *api.APIError
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
			mockSave: func(id, metricType, rawValue string) *api.APIError {
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
			mockSave: func(id, metricType, rawValue string) *api.APIError {
				return api.BadRequest("invalid metric value")
			},
			expectStatus:   http.StatusBadRequest,
			expectErrorMsg: "invalid metric value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := service.NewMockMetricsService(t)
			mockService.EXPECT().
				Save(tt.pathVals["id"], tt.pathVals["type"], tt.pathVals["value"]).
				RunAndReturn(tt.mockSave)

			handler, err := NewMetricsHandler(mockService)
			assert.NoError(t, err)

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
		mockGet        func(metricID, metricType string) (any, *api.APIError)
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
			mockGet: func(metricID, metricType string) (any, *api.APIError) {
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
			mockGet: func(metricID, metricType string) (any, *api.APIError) {
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
			mockGet: func(metricID, metricType string) (any, *api.APIError) {
				return nil, api.NotFound("metric m2 not found")
			},
			expectStatus:   http.StatusNotFound,
			expectErrorMsg: "metric m2 not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := service.NewMockMetricsService(t)
			mockService.EXPECT().
				Get(tt.pathVals["id"], tt.pathVals["type"]).
				RunAndReturn(tt.mockGet)

			handler, err := NewMetricsHandler(mockService)
			assert.NoError(t, err)

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
	}{
		{
			name:           "nil metrics",
			mockReturn:     nil,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "empty metrics",
			mockReturn:     &map[string]any{},
			expectedStatus: http.StatusOK,
		},
		{
			name: "metrics with values",
			mockReturn: &map[string]any{
				"c1": int64(10),
				"g1": float64(3.14),
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := service.NewMockMetricsService(t)
			mockService.EXPECT().GetAll().Return(tt.mockReturn)

			handler, err := NewMetricsHandler(mockService)
			assert.NoError(t, err)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()

			handler.GetAll(rec, req)

			res := rec.Result()
			defer res.Body.Close()

			body, _ := io.ReadAll(res.Body)
			assert.Equal(t, tt.expectedStatus, res.StatusCode)

			bodyStr := string(body)
			if tt.mockReturn == nil {
				assert.Contains(t, bodyStr, "Metrics not found")
			} else {
				for id, val := range *tt.mockReturn {
					assert.Contains(t, bodyStr, id)
					assert.Contains(t, bodyStr, fmt.Sprintf("%v", val))
				}
				assert.Contains(t, bodyStr, "<h1>Metrics</h1>")
			}
		})
	}
}

func strPtr(value string) *string {
	return &value
}
