package handler

import (
	"encoding/json"
	"errors"
	"html/template"
	"net/http"

	models "github.com/gabkaclassic/metrics/internal/model"
	"github.com/gabkaclassic/metrics/internal/service"
	api "github.com/gabkaclassic/metrics/pkg/error"
)

type MetricsPageData struct {
	Metrics map[string]any
}

var metricsTemplate = template.Must(template.New("metrics").Parse(`
<html>
<head>
	<title>Metrics</title>
</head>
<body>
	<h1>Metrics</h1>
	<table border="1" cellpadding="5" cellspacing="0">
		<tr><th>ID</th><th>Value</th></tr>
		{{range $id, $val := .Metrics}}
		<tr>
			<td>{{$id}}</td>
			<td>{{$val}}</td>
		</tr>
		{{end}}
	</table>
</body>
</html>
`))

type MetricsHandler struct {
	service service.MetricsService
}

func NewMetricsHandler(service service.MetricsService) (*MetricsHandler, error) {

	if service == nil {
		return nil, errors.New("create new metrics handler failed: service is nil")
	}

	return &MetricsHandler{
		service: service,
	}, nil
}

// Save saves a single metric using plain-text URL parameters.
//
// @Summary Save metric (plain-text)
// @Description Saves a metric using URL parameters. Counters are incremented, gauges are overwritten.
// @Tags Metrics
// @Param type path string true "Metric type" Enums(gauge,counter)
// @Param id path string true "Metric ID"
// @Param value path string true "Metric value"
// @Success 200 "Metric saved"
// @Failure 400 {object} api.APIError "Bad Request"
// @Failure 404 {object} api.APIError "Not Found"
// @Failure 500 {object} api.APIError "Internal Error"
// @Router /update/{type}/{id}/{value} [post]
func (handler *MetricsHandler) Save(w http.ResponseWriter, r *http.Request) {

	metricID := r.PathValue("id")
	metricType := r.PathValue("type")
	metricValue := r.PathValue("value")

	err := handler.service.Save(r.Context(), metricID, metricType, metricValue)

	if err != nil {
		api.RespondError(w, err)
		return
	}
}

// SaveJSON saves a single metric using JSON payload.
//
// @Summary Save metric (JSON)
// @Description Saves a metric using JSON body. Counters are incremented, gauges are overwritten.
// @Tags Metrics
// @Accept json
// @Produce json
// @Param metric body models.Metrics true "Metric payload"
// @Success 200 {object} models.Metrics "Saved metric"
// @Failure 400 {object} api.APIError "Bad Request"
// @Failure 422 {object} api.APIError "Invalid JSON"
// @Failure 500 {object} api.APIError "Internal Error"
// @Router /update [post]
func (handler *MetricsHandler) SaveJSON(w http.ResponseWriter, r *http.Request) {
	metric := &models.Metrics{}
	err := json.NewDecoder(r.Body).Decode(metric)
	if err != nil {
		api.RespondError(w, api.UnprocessibleEntity("Invalid input JSON"))
		return
	}

	saveErr := handler.service.SaveStruct(r.Context(), *metric)

	if saveErr != nil {
		api.RespondError(w, saveErr)
		return
	}

	encodeErr := json.NewEncoder(w).Encode(metric)

	if encodeErr != nil {
		api.RespondError(w, encodeErr)
		return
	}
}

// SaveAll saves multiple metrics in a single request.
//
// @Summary Save metrics batch
// @Description Saves multiple metrics. Counters are aggregated by ID, gauges use the last value.
// @Tags Metrics
// @Accept json
// @Param metrics body []models.Metrics true "Metrics list"
// @Success 200 "Metrics saved"
// @Failure 400 {object} api.APIError "Bad Request"
// @Failure 422 {object} api.APIError "Invalid JSON"
// @Failure 500 {object} api.APIError "Internal Error"
// @Router /updates [post]
func (handler *MetricsHandler) SaveAll(w http.ResponseWriter, r *http.Request) {
	metrics := make([]models.Metrics, 0)
	err := json.NewDecoder(r.Body).Decode(&metrics)
	if err != nil {
		api.RespondError(w, api.UnprocessibleEntity("Invalid input JSON"))
		return
	}

	saveErr := handler.service.SaveAll(r.Context(), metrics)

	if saveErr != nil {
		api.RespondError(w, saveErr)
		return
	}
}

// Get retrieves a metric value by ID and type.
//
// @Summary Get metric value
// @Description Returns raw metric value. Counter → int64, Gauge → float64.
// @Tags Metrics
// @Produce json
// @Param type path string true "Metric type" Enums(gauge,counter)
// @Param id path string true "Metric ID"
// @Success 200 {object} any "Metric value"
// @Failure 404 {object} api.APIError "Not Found"
// @Failure 500 {object} api.APIError "Internal Error"
// @Router /value/{type}/{id} [get]
func (handler *MetricsHandler) Get(w http.ResponseWriter, r *http.Request) {

	metricID := r.PathValue("id")
	metricType := r.PathValue("type")

	value, err := handler.service.Get(r.Context(), metricID, metricType)

	if err != nil {
		api.RespondError(w, err)
		return
	}

	encodeErr := json.NewEncoder(w).Encode(value)

	if encodeErr != nil {
		api.RespondError(w, encodeErr)
		return
	}
}

// GetJSON retrieves a metric using JSON request.
//
// @Summary Get metric (JSON)
// @Description Returns full metric structure.
// @Tags Metrics
// @Accept json
// @Produce json
// @Param metric body models.Metrics true "Metric identifier"
// @Success 200 {object} models.Metrics "Metric data"
// @Failure 404 {object} api.APIError "Not Found"
// @Failure 422 {object} api.APIError "Invalid JSON"
// @Failure 500 {object} api.APIError "Internal Error"
// @Router /value [post]
func (handler *MetricsHandler) GetJSON(w http.ResponseWriter, r *http.Request) {

	metric := &models.Metrics{}
	err := json.NewDecoder(r.Body).Decode(metric)

	if err != nil {
		api.RespondError(w, api.UnprocessibleEntity("Invalid input JSON"))
		return
	}

	value, getErr := handler.service.GetStruct(r.Context(), metric.ID, metric.MType)

	if getErr != nil {
		api.RespondError(w, getErr)
		return
	}

	encodeErr := json.NewEncoder(w).Encode(value)

	if encodeErr != nil {
		api.RespondError(w, encodeErr)
		return
	}
}

// GetAll renders all metrics as an HTML page.
//
// @Summary Get all metrics (HTML)
// @Description Returns all stored metrics rendered as HTML table.
// @Tags Metrics
// @Produce text/html
// @Success 200 "HTML page with metrics"
// @Failure 404 {object} api.APIError "Not Found"
// @Failure 500 {object} api.APIError "Internal Error"
// @Router / [get]
func (handler *MetricsHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	metrics, err := handler.service.GetAll(r.Context())

	if err != nil {
		api.RespondError(w, err)
		return
	}

	if metrics == nil {
		api.RespondError(
			w,
			api.NotFound("Metrics not found"),
		)
		return
	}

	data := MetricsPageData{
		Metrics: metrics,
	}

	if err := metricsTemplate.Execute(w, data); err != nil {
		api.RespondError(
			w,
			api.Internal("failed to render template", err),
		)
		return
	}
}
