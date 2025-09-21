package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	api "github.com/gabkaclassic/metrics/internal/error"
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

	metricID := r.PathValue("id")
	metricType := r.PathValue("type")
	metricValue := r.PathValue("value")

	err := handler.service.Save(metricID, metricType, metricValue)

	if err != nil {
		api.RespondError(w, err)
		return
	}
}

func (handler *MetricsHandler) Get(w http.ResponseWriter, r *http.Request) {

	metricID := r.PathValue("id")
	metricType := r.PathValue("type")

	value, err := handler.service.Get(metricID, metricType)

	if err != nil {
		api.RespondError(w, err)
		return
	}

	json.NewEncoder(w).Encode(value)
}

func (handler *MetricsHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	metrics := handler.service.GetAll()
	if metrics == nil {
		api.RespondError(
			w,
			api.NotFound("Metrics not found"),
		)
		return
	}

	builder := strings.Builder{}
	builder.WriteString("<html><head><title>Metrics</title></head><body>")
	builder.WriteString("<h1>Metrics</h1>")
	builder.WriteString("<table border='1' cellpadding='5' cellspacing='0'>")
	builder.WriteString("<tr><th>ID</th><th>Value</th></tr>")

	for id, val := range *metrics {
		builder.WriteString("<tr>")
		builder.WriteString("<td>" + id + "</td>")
		builder.WriteString("<td>" + fmt.Sprintf("%v", val) + "</td>")
		builder.WriteString("</tr>")
	}

	builder.WriteString("</table>")
	builder.WriteString("</body></html>")

	_, _ = w.Write([]byte(builder.String()))
}
