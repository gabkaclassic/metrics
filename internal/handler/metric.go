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

func (handler *MetricsHandler) SaveJSON(w http.ResponseWriter, r *http.Request) {
	metric := &models.Metrics{}
	err := json.NewDecoder(r.Body).Decode(metric)
	if err != nil {
		api.RespondError(w, api.UnprocessibleEntity("Invalid input JSON"))
		return
	}

	saveErr := handler.service.SaveStruct(*metric)

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

func (handler *MetricsHandler) Get(w http.ResponseWriter, r *http.Request) {

	metricID := r.PathValue("id")
	metricType := r.PathValue("type")

	value, err := handler.service.Get(metricID, metricType)

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

func (handler *MetricsHandler) GetJSON(w http.ResponseWriter, r *http.Request) {

	metric := &models.Metrics{}
	err := json.NewDecoder(r.Body).Decode(metric)

	if err != nil {
		api.RespondError(w, api.UnprocessibleEntity("Invalid input JSON"))
		return
	}

	value, getErr := handler.service.GetStruct(metric.ID, metric.MType)

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

func (handler *MetricsHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	metrics, err := handler.service.GetAll()

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
		Metrics: *metrics,
	}

	if err := metricsTemplate.Execute(w, data); err != nil {
		api.RespondError(
			w,
			api.Internal("failed to render template", err),
		)
		return
	}
}
