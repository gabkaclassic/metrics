package service

import (
	"errors"
	"fmt"
	"strconv"

	models "github.com/gabkaclassic/metrics/internal/model"
	"github.com/gabkaclassic/metrics/internal/repository"
	api "github.com/gabkaclassic/metrics/pkg/error"
)

type MetricsService interface {
	Get(metricID string, metricType string) (any, *api.APIError)
	GetStruct(metricID string, metricType string) (*models.Metrics, *api.APIError)
	Save(id string, metricType string, rawValue string) *api.APIError
	SaveStruct(metric models.Metrics) *api.APIError
	SaveAll(metrics *[]models.Metrics) *api.APIError
	GetAll() (*map[string]any, *api.APIError)
}

type metricsService struct {
	repository repository.MetricsRepository
}

func NewMetricsService(repository repository.MetricsRepository) (MetricsService, error) {

	if repository == nil {
		return nil, errors.New("create new metrics service failed: repository is nil")
	}

	return &metricsService{
		repository: repository,
	}, nil
}

func (service *metricsService) GetAll() (*map[string]any, *api.APIError) {
	metrics, err := service.repository.GetAll()

	if err != nil {
		return nil, api.Internal("Get all metrics error", err)
	}

	return metrics, nil
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

func (service *metricsService) GetStruct(metricID string, metricType string) (*models.Metrics, *api.APIError) {

	metric, err := service.repository.Get(metricID)

	if metric == nil || metric.MType != metricType {
		return nil, api.NotFound(fmt.Sprintf("metric %v %v not found", metricID, metricType))
	}

	if err != nil {
		return nil, api.Internal("Get metric error", err)
	}

	return &models.Metrics{
		ID:    metricID,
		MType: metricType,
		Value: metric.Value,
		Delta: metric.Delta,
	}, nil
}

func (service *metricsService) Save(id string, metricType string, rawValue string) *api.APIError {

	switch metricType {
	case models.Counter:
		if delta, err := strconv.ParseInt(rawValue, 10, 64); err == nil {
			err := service.repository.Add(models.Metrics{
				ID:    id,
				MType: metricType,
				Delta: &delta,
			})
			if err != nil {
				return api.Internal("Add delta error", err)
			}
		} else {
			return api.BadRequest(fmt.Sprintf("invalid metric value: %s", rawValue))
		}
	case models.Gauge:
		if value, err := strconv.ParseFloat(rawValue, 64); err == nil {
			err := service.repository.Reset(models.Metrics{
				ID:    id,
				MType: models.Gauge,
				Value: &value,
			})
			if err != nil {
				return api.Internal("Reset value error", err)
			}
		} else {
			return api.BadRequest(fmt.Sprintf("invalid metric value: %s", rawValue))
		}
	default:
		return api.BadRequest(fmt.Sprintf("invalid metric type: %s", metricType))
	}

	return nil
}

func (service *metricsService) SaveStruct(metric models.Metrics) *api.APIError {

	switch metric.MType {
	case models.Counter:
		service.repository.Add(metric)
	case models.Gauge:
		service.repository.Reset(metric)
	default:
		return api.BadRequest(fmt.Sprintf("invalid metric type: %s", metric.MType))
	}

	return nil
}

func (service *metricsService) SaveAll(metrics *[]models.Metrics) *api.APIError {
	counterSums := make(map[string]int64)
	gaugeLastValues := make(map[string]float64)

	for _, metric := range *metrics {
		switch metric.MType {
		case models.Counter:
			if metric.Delta != nil {
				counterSums[metric.ID] += *metric.Delta
			}
		case models.Gauge:
			if metric.Value != nil {
				gaugeLastValues[metric.ID] = *metric.Value
			}
		default:
			return api.BadRequest(fmt.Sprintf("invalid metric type: %s", metric.MType))
		}
	}

	counters := make([]models.Metrics, 0, len(counterSums))
	for id, delta := range counterSums {
		deltaCopy := delta
		counters = append(counters, models.Metrics{
			ID:    id,
			MType: models.Counter,
			Delta: &deltaCopy,
		})
	}

	gauges := make([]models.Metrics, 0, len(gaugeLastValues))
	for id, value := range gaugeLastValues {
		valueCopy := value
		gauges = append(gauges, models.Metrics{
			ID:    id,
			MType: models.Gauge,
			Value: &valueCopy,
		})
	}

	errChan := make(chan error, 2)

	if len(counters) > 0 {
		go func() { errChan <- service.repository.AddAll(&counters) }()
	} else {
		go func() { errChan <- nil }()
	}

	if len(gauges) > 0 {
		go func() { errChan <- service.repository.ResetAll(&gauges) }()
	} else {
		go func() { errChan <- nil }()
	}

	err1 := <-errChan
	err2 := <-errChan

	if err1 != nil || err2 != nil {
		return api.Internal("save metrics error", fmt.Errorf("counters: %v, gauges: %v", err1, err2))
	}

	return nil
}
