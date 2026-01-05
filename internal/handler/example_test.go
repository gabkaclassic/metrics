package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	models "github.com/gabkaclassic/metrics/internal/model"
	api "github.com/gabkaclassic/metrics/pkg/error"
)

type stubService struct{}

func (s *stubService) Save(ctx context.Context, id, mtype, value string) *api.APIError {
	return nil
}
func (s *stubService) SaveAll(ctx context.Context, metrics []models.Metrics) *api.APIError {
	return nil
}
func (s *stubService) SaveStruct(ctx context.Context, m models.Metrics) *api.APIError { return nil }
func (s *stubService) Get(ctx context.Context, id, mtype string) (any, *api.APIError) { return 42, nil }
func (s *stubService) GetStruct(ctx context.Context, id, mtype string) (models.Metrics, *api.APIError) {
	return models.Metrics{ID: id, MType: mtype}, nil
}
func (s *stubService) GetAll(ctx context.Context) (map[string]any, *api.APIError) {
	return map[string]any{"m1": floatPtr(1.23)}, nil
}

// ExampleMetricsHandler_Save shows how to call the Save endpoint (plain-text).
func ExampleMetricsHandler_Save() {
	svc := &stubService{}
	h, _ := NewMetricsHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/update/counter/mycounter/42", nil)
	w := httptest.NewRecorder()

	h.Save(w, req)

	fmt.Println(w.Code)
	// Output: 200
}

// ExampleMetricsHandler_SaveJSON shows how to call the SaveJSON endpoint.
func ExampleMetricsHandler_SaveJSON() {
	svc := &stubService{}
	h, _ := NewMetricsHandler(svc)

	metric := models.Metrics{
		ID:    "mygauge",
		MType: "gauge",
		Value: floatPtr(3.14),
	}
	body, _ := json.Marshal(metric)

	req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.SaveJSON(w, req)

	fmt.Println(w.Code)
	// Output: 200
}

// ExampleMetricsHandler_Get shows how to call the Get endpoint.
func ExampleMetricsHandler_Get() {
	svc := &stubService{}
	h, _ := NewMetricsHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/value/counter/mycounter", nil)
	w := httptest.NewRecorder()

	h.Get(w, req)

	fmt.Println(w.Code)
	// Output: 200
}

// ExampleMetricsHandler_GetJSON shows how to call the GetJSON endpoint.
func ExampleMetricsHandler_GetJSON() {
	svc := &stubService{}
	h, _ := NewMetricsHandler(svc)

	metric := models.Metrics{
		ID:    "mygauge",
		MType: "gauge",
	}
	body, _ := json.Marshal(metric)

	req := httptest.NewRequest(http.MethodPost, "/value", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.GetJSON(w, req)

	fmt.Println(w.Code)
	// Output: 200
}

// ExampleMetricsHandler_GetAll shows how to call the GetAll endpoint.
func ExampleMetricsHandler_GetAll() {
	svc := &stubService{}
	h, _ := NewMetricsHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	h.GetAll(w, req)

	fmt.Println(w.Code)
	// Output: 200
}

func floatPtr(f float64) *float64 { return &f }
