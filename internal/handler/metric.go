package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	api_error "github.com/gabkaclassic/metrics/internal/error"
	"github.com/gabkaclassic/metrics/internal/service"
)

type MetricsHandler struct {
	service service.MetricsService
}

func NewMetricsHandler(service service.MetricsService) *MetricsHandler {

	if service == nil {
		panic(errors.New("create new metrics handler failed: service is nil"))
	}

	return &MetricsHandler{
		service: service,
	}
}

func (handler *MetricsHandler) Save(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		api_error.RespondError(w, api_error.NotAllowed())
		return
	}

	metricID := r.PathValue("id")
	metricType := r.PathValue("type")
	metricValue := r.PathValue("value")

	err := handler.service.Save(metricID, metricType, metricValue)

	if err != nil {
		api_error.RespondError(w, err)
		return
	}
}

func (handler *MetricsHandler) Get(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		api_error.RespondError(w, api_error.NotAllowed())
		return
	}

	metricID := r.PathValue("id")

	metric, err := handler.service.Get(metricID)

	if err != nil {
		api_error.RespondError(w, err)
		return
	}

	json.NewEncoder(w).Encode(metric)
}
