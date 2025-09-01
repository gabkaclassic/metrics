package service

import (
	"errors"
	"fmt"
	"strconv"

	api_error "github.com/gabkaclassic/metrics/internal/error"
	models "github.com/gabkaclassic/metrics/internal/model"
	"github.com/gabkaclassic/metrics/internal/repository"
)

type MetricsService struct {
	repository repository.MetricsRepository
}

func NewMetricsService(repository repository.MetricsRepository) *MetricsService {

	if repository == nil {
		panic(errors.New("create new metrics service failed: repository is nil"))
	}

	return &MetricsService{
		repository: repository,
	}
}

func (service *MetricsService) SaveMetric(id string, metricType string, rawValue string) *api_error.ApiError {

	switch metricType {
	case models.Counter:
		if delta, err := strconv.ParseInt(rawValue, 10, 64); err == nil {
			service.repository.Add(models.Metrics{
				ID:    id,
				MType: metricType,
				Delta: &delta,
			})
		} else {
			return api_error.BadRequest(fmt.Sprintf("invalid metric value: %s", rawValue))
		}
	case models.Gauge:
		if value, err := strconv.ParseFloat(rawValue, 64); err == nil {
			service.repository.Reset(models.Metrics{
				ID:    id,
				MType: models.Gauge,
				Value: &value,
			})
		} else {
			return api_error.BadRequest(fmt.Sprintf("invalid metric value: %s", rawValue))
		}
	default:
		return api_error.BadRequest(fmt.Sprintf("invalid metric type: %s", metricType))
	}

	return nil
}
