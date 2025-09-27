package handler

import (
	"encoding/json"
	"errors"
	"html/template"
	"net/http"

	api "github.com/gabkaclassic/metrics/internal/error"
	"github.com/gabkaclassic/metrics/internal/service"
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

	data := MetricsPageData{
		Metrics: *metrics,
	}

	if err := metricsTemplate.Execute(w, data); err != nil {
		http.Error(w, "failed to render template", http.StatusInternalServerError)
		return
	}
}
