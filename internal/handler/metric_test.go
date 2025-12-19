package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"

	models "github.com/gabkaclassic/metrics/internal/model"
	"github.com/gabkaclassic/metrics/internal/service"
	"github.com/stretchr/testify/mock"

	"fmt"
	"io"
	"strings"
	"testing"

	api "github.com/gabkaclassic/metrics/pkg/error"
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
		mockSave       func(ctx context.Context, id, metricType, rawValue string) *api.APIError
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
			mockSave: func(ctx context.Context, id, metricType, rawValue string) *api.APIError {
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
			mockSave: func(ctx context.Context, id, metricType, rawValue string) *api.APIError {
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
				Save(mock.Anything, tt.pathVals["id"], tt.pathVals["type"], tt.pathVals["value"]).
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

func TestMetricsHandler_SaveAll(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		requestBody    string
		mockSaveAll    func(ctx context.Context, metrics []models.Metrics) *api.APIError
		expectStatus   int
		expectErrorMsg string
	}{
		{
			name:   "valid POST with counter metrics",
			method: http.MethodPost,
			requestBody: `[
				{"id": "c1", "type": "counter", "delta": 10},
				{"id": "c2", "type": "counter", "delta": 5}
			]`,
			mockSaveAll: func(ctx context.Context, metrics []models.Metrics) *api.APIError {
				assert.Len(t, metrics, 2)
				assert.Equal(t, "c1", (metrics)[0].ID)
				assert.Equal(t, models.Counter, (metrics)[0].MType)
				assert.Equal(t, int64(10), *(metrics)[0].Delta)
				return nil
			},
			expectStatus: http.StatusOK,
		},
		{
			name:   "valid POST with gauge metrics",
			method: http.MethodPost,
			requestBody: `[
				{"id": "g1", "type": "gauge", "value": 3.14},
				{"id": "g2", "type": "gauge", "value": 2.71}
			]`,
			mockSaveAll: func(ctx context.Context, metrics []models.Metrics) *api.APIError {
				assert.Len(t, metrics, 2)
				assert.Equal(t, "g1", (metrics)[0].ID)
				assert.Equal(t, models.Gauge, (metrics)[0].MType)
				assert.Equal(t, 3.14, *(metrics)[0].Value)
				return nil
			},
			expectStatus: http.StatusOK,
		},
		{
			name:   "valid POST with mixed metrics",
			method: http.MethodPost,
			requestBody: `[
				{"id": "c1", "type": "counter", "delta": 10},
				{"id": "g1", "type": "gauge", "value": 3.14}
			]`,
			mockSaveAll: func(ctx context.Context, metrics []models.Metrics) *api.APIError {
				assert.Len(t, metrics, 2)
				return nil
			},
			expectStatus: http.StatusOK,
		},
		{
			name:        "empty metrics array",
			method:      http.MethodPost,
			requestBody: `[]`,
			mockSaveAll: func(ctx context.Context, metrics []models.Metrics) *api.APIError {
				assert.Len(t, metrics, 0)
				return nil
			},
			expectStatus: http.StatusOK,
		},
		{
			name:           "invalid JSON",
			method:         http.MethodPost,
			requestBody:    `invalid json`,
			mockSaveAll:    func(ctx context.Context, metrics []models.Metrics) *api.APIError { return nil },
			expectStatus:   http.StatusUnprocessableEntity,
			expectErrorMsg: "Invalid input JSON",
		},
		{
			name:   "service returns error",
			method: http.MethodPost,
			requestBody: `[
				{"id": "c1", "type": "counter", "delta": 10}
			]`,
			mockSaveAll: func(ctx context.Context, metrics []models.Metrics) *api.APIError {
				return api.Internal("save error", errors.New("some error"))
			},
			expectStatus:   http.StatusInternalServerError,
			expectErrorMsg: "save error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := service.NewMockMetricsService(t)
			if tt.expectStatus != http.StatusUnprocessableEntity {
				mockService.EXPECT().
					SaveAll(mock.Anything, mock.Anything).
					RunAndReturn(tt.mockSaveAll)
			}

			handler, err := NewMetricsHandler(mockService)
			assert.NoError(t, err)

			req := httptest.NewRequest(tt.method, "/", strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			handler.SaveAll(rr, req)

			assert.Equal(t, tt.expectStatus, rr.Code)
			if tt.expectErrorMsg != "" {
				assert.Contains(t, rr.Body.String(), tt.expectErrorMsg)
			}
		})
	}
}

func TestMetricsHandler_SaveJSON(t *testing.T) {
	tests := []struct {
		name              string
		body              string
		mockSave          func(ctx context.Context, metric models.Metrics) *api.APIError
		expectStatus      int
		expectServiceCall bool
		expectErrorMsg    string
	}{
		{
			name: "valid JSON",
			body: `{"id":"m1","type":"counter","delta":10}`,
			mockSave: func(ctx context.Context, metric models.Metrics) *api.APIError {
				assert.Equal(t, "m1", metric.ID)
				assert.Equal(t, "counter", metric.MType)
				assert.Equal(t, int64(10), *metric.Delta)
				return nil
			},
			expectStatus:      http.StatusOK,
			expectServiceCall: true,
		},
		{
			name:           "invalid JSON",
			body:           `{"id": "m1", "type":`,
			mockSave:       func(ctx context.Context, metric models.Metrics) *api.APIError { return nil },
			expectStatus:   http.StatusUnprocessableEntity,
			expectErrorMsg: "Invalid input JSON",
		},
		{
			name: "service error",
			body: `{"id":"m2","type":"gauge","value":3.14}`,
			mockSave: func(ctx context.Context, metric models.Metrics) *api.APIError {
				return api.BadRequest("save failed")
			},
			expectStatus:      http.StatusBadRequest,
			expectServiceCall: true,
			expectErrorMsg:    "save failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := service.NewMockMetricsService(t)

			if tt.expectServiceCall {
				mockService.EXPECT().
					SaveStruct(mock.Anything, mock.AnythingOfType("models.Metrics")).
					RunAndReturn(tt.mockSave)
			}

			handler, err := NewMetricsHandler(mockService)
			assert.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.body))
			rr := httptest.NewRecorder()

			handler.SaveJSON(rr, req)

			assert.Equal(t, tt.expectStatus, rr.Code)
			if tt.expectErrorMsg != "" {
				assert.Contains(t, rr.Body.String(), tt.expectErrorMsg)
			}
		})
	}
}

func TestMetricsHandler_GetJSON(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		mockGet        func(ctx context.Context, id, mType string) (models.Metrics, *api.APIError)
		expectStatus   int
		expectErrorMsg string
		expectGetCall  bool
	}{
		{
			name: "valid JSON and service success",
			body: `{"id":"m1","type":"counter"}`,
			mockGet: func(ctx context.Context, id, mType string) (models.Metrics, *api.APIError) {
				assert.Equal(t, "m1", id)
				assert.Equal(t, "counter", mType)
				delta := int64(42)
				return models.Metrics{
					ID:    id,
					MType: mType,
					Delta: &delta,
				}, nil
			},
			expectStatus:  http.StatusOK,
			expectGetCall: true,
		},
		{
			name:           "invalid JSON",
			body:           `{"id":"m1","type":`,
			expectStatus:   http.StatusUnprocessableEntity,
			expectErrorMsg: "Invalid input JSON",
			expectGetCall:  false,
		},
		{
			name: "service returns error",
			body: `{"id":"m2","type":"gauge"}`,
			mockGet: func(ctx context.Context, id, mType string) (models.Metrics, *api.APIError) {
				return models.Metrics{}, api.NotFound("metric not found")
			},
			expectStatus:   http.StatusNotFound,
			expectErrorMsg: "metric not found",
			expectGetCall:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := service.NewMockMetricsService(t)

			if tt.expectGetCall {
				mockService.EXPECT().
					GetStruct(mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).
					RunAndReturn(tt.mockGet)
			}

			handler, err := NewMetricsHandler(mockService)
			assert.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.body))
			rr := httptest.NewRecorder()

			handler.GetJSON(rr, req)

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
		mockGet        func(ctx context.Context, metricID, metricType string) (any, *api.APIError)
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
			mockGet: func(ctx context.Context, metricID, metricType string) (any, *api.APIError) {
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
			mockGet: func(ctx context.Context, metricID, metricType string) (any, *api.APIError) {
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
			mockGet: func(ctx context.Context, metricID, metricType string) (any, *api.APIError) {
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
				Get(mock.Anything, tt.pathVals["id"], tt.pathVals["type"]).
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
		mockReturn     map[string]any
		expectedStatus int
		expectedError  *api.APIError
	}{
		{
			name:           "nil metrics",
			mockReturn:     nil,
			expectedStatus: http.StatusNotFound,
			expectedError:  nil,
		},
		{
			name:           "empty metrics",
			mockReturn:     map[string]any{},
			expectedStatus: http.StatusOK,
			expectedError:  nil,
		},
		{
			name: "metrics with values",
			mockReturn: map[string]any{
				"c1": int64(10),
				"g1": float64(3.14),
			},
			expectedStatus: http.StatusOK,
			expectedError:  nil,
		},
		{
			name:           "return error",
			mockReturn:     nil,
			expectedStatus: http.StatusInternalServerError,
			expectedError:  api.Internal("some error", errors.New("some error")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := service.NewMockMetricsService(t)
			mockService.EXPECT().GetAll(mock.Anything).Return(tt.mockReturn, tt.expectedError)

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
				if tt.expectedError == nil {
					assert.Contains(t, bodyStr, "Metrics not found")
				}
			} else {
				for id, val := range tt.mockReturn {
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
