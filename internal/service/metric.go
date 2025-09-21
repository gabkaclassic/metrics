package service

import (
	"errors"
	"fmt"
	"strconv"

	api "github.com/gabkaclassic/metrics/internal/error"
	models "github.com/gabkaclassic/metrics/internal/model"
	"github.com/gabkaclassic/metrics/internal/repository"
)

type MetricsService interface {
	Get(metricID string, metricType string) (any, *api.APIError)
	Save(id string, metricType string, rawValue string) *api.APIError
	GetAll() *map[string]any
}

type metricsService struct {
	repository repository.MetricsRepository
}

func NewMetricsService(repository repository.MetricsRepository) MetricsService {

	if repository == nil {
		panic(errors.New("create new metrics service failed: repository is nil"))
	}

	return &metricsService{
		repository: repository,
	}
}

func (service *metricsService) GetAll() *map[string]any {
	return service.repository.GetAll()
}

func (service *metricsService) Get(metricID string, metricType string) (any, *api.APIError) {

	metric, err := service.repository.Get(metricID)

	if metric == nil || metric.MType != metricType {
		return nil, api.NotFound(fmt.Sprintf("Metric %s with type %s not found", metricID, metricType))
	}

	if err != nil {
		return nil, api.Internal("Get metric error", err)
	}

	switch metric.MType {
	case models.Counter:
		return metric.Delta, nil
	case models.Gauge:
		return metric.Value, nil
	default:
		return nil, api.BadRequest(fmt.Sprintf("Unknown metric type: %s", metricType))
	}
}

func (service *metricsService) Save(id string, metricType string, rawValue string) *api.APIError {

	switch metricType {
	case models.Counter:
		if delta, err := strconv.ParseInt(rawValue, 10, 64); err == nil {
			service.repository.Add(models.Metrics{
				ID:    id,
				MType: metricType,
				Delta: &delta,
			})
		} else {
			return api.BadRequest(fmt.Sprintf("invalid metric value: %s", rawValue))
		}
	case models.Gauge:
		if value, err := strconv.ParseFloat(rawValue, 64); err == nil {
			service.repository.Reset(models.Metrics{
				ID:    id,
				MType: models.Gauge,
				Value: &value,
			})
		} else {
			return api.BadRequest(fmt.Sprintf("invalid metric value: %s", rawValue))
		}
	default:
		return api.BadRequest(fmt.Sprintf("invalid metric type: %s", metricType))
	}

	return nil
}
