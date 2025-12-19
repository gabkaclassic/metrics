package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/gabkaclassic/metrics/internal/audit"
	models "github.com/gabkaclassic/metrics/internal/model"
	"github.com/gabkaclassic/metrics/internal/repository"
	api "github.com/gabkaclassic/metrics/pkg/error"
	"github.com/gabkaclassic/metrics/pkg/middleware"
)

type MetricsService interface {
	Get(context.Context, string, string) (any, *api.APIError)
	GetStruct(context.Context, string, string) (models.Metrics, *api.APIError)
	Save(context.Context, string, string, string) *api.APIError
	SaveStruct(context.Context, models.Metrics) *api.APIError
	SaveAll(context.Context, []models.Metrics) *api.APIError
	GetAll(context.Context) (map[string]any, *api.APIError)
}

type metricsService struct {
	repository repository.MetricsRepository
	auditor    audit.Auditor
}

func NewMetricsService(repository repository.MetricsRepository, auditor audit.Auditor) (MetricsService, error) {

	if repository == nil {
		return nil, errors.New("create new metrics service failed: repository is nil")
	}

	if auditor == nil {
		return nil, errors.New("create new metrics service failed: auditor is nil")
	}

	return &metricsService{
		repository: repository,
		auditor:    auditor,
	}, nil
}

func (service *metricsService) notifyOne(ctx context.Context, metric models.Metrics) {

	ts := middleware.AuditTsFromCtx(ctx)

	if ts == 0 {
		slog.Error("Get audit timestamp from request context error")
		return
	}

	ip := middleware.AuditIPFromCtx(ctx)

	if len(ip) == 0 {
		slog.Error("Get audit source IP from request context error")
		return
	}

	service.auditor.AuditOne(metric, ts, ip)
}

func (service *metricsService) notifyMany(ctx context.Context, metrics []models.Metrics) {

	if len(metrics) == 0 {
		slog.Debug("Metrics list for audit is empty")
		return
	}

	ts := middleware.AuditTsFromCtx(ctx)

	if ts == 0 {
		slog.Error("Get audit timestamp from request context error")
		return
	}

	ip := middleware.AuditIPFromCtx(ctx)

	if len(ip) == 0 {
		slog.Error("Get audit source IP from request context error")
		return
	}

	service.auditor.AuditMany(metrics, ts, ip)
}

func (service *metricsService) GetAll(ctx context.Context) (map[string]any, *api.APIError) {
	metrics, err := service.repository.GetAll(ctx)

	if err != nil {
		return nil, api.Internal("Get all metrics error", err)
	}

	return metrics, nil
}

func (service *metricsService) Get(ctx context.Context, metricID string, metricType string) (any, *api.APIError) {

	metric, err := service.repository.Get(ctx, metricID)

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

func (service *metricsService) GetStruct(ctx context.Context, metricID string, metricType string) (models.Metrics, *api.APIError) {

	metric, err := service.repository.Get(ctx, metricID)

	if metric == nil || metric.MType != metricType {
		return models.Metrics{}, api.NotFound(fmt.Sprintf("metric %v %v not found", metricID, metricType))
	}

	if err != nil {
		return models.Metrics{}, api.Internal("Get metric error", err)
	}

	return models.Metrics{
		ID:    metricID,
		MType: metricType,
		Value: metric.Value,
		Delta: metric.Delta,
	}, nil
}

func (service *metricsService) Save(ctx context.Context, id string, metricType string, rawValue string) *api.APIError {

	switch metricType {
	case models.Counter:
		if delta, err := strconv.ParseInt(rawValue, 10, 64); err == nil {
			metric := models.Metrics{
				ID:    id,
				MType: metricType,
				Delta: &delta,
			}
			err := service.repository.Add(ctx, metric)
			if err != nil {
				return api.Internal("Add delta error", err)
			}
			go service.notifyOne(ctx, metric)
		} else {
			return api.BadRequest(fmt.Sprintf("invalid metric value: %s", rawValue))
		}
	case models.Gauge:
		if value, err := strconv.ParseFloat(rawValue, 64); err == nil {
			metric := models.Metrics{
				ID:    id,
				MType: models.Gauge,
				Value: &value,
			}
			err := service.repository.Reset(ctx, metric)
			if err != nil {
				return api.Internal("Reset value error", err)
			}
			go service.notifyOne(ctx, metric)
		} else {
			return api.BadRequest(fmt.Sprintf("invalid metric value: %s", rawValue))
		}
	default:
		return api.BadRequest(fmt.Sprintf("invalid metric type: %s", metricType))
	}

	return nil
}

func (service *metricsService) SaveStruct(ctx context.Context, metric models.Metrics) *api.APIError {

	var err error
	switch metric.MType {
	case models.Counter:
		err = service.repository.Add(ctx, metric)
	case models.Gauge:
		err = service.repository.Reset(ctx, metric)
	default:
		return api.BadRequest(fmt.Sprintf("invalid metric type: %s", metric.MType))
	}

	if err != nil {
		return api.Internal("save metric error", err)
	}
	go service.notifyOne(ctx, metric)

	return nil
}

func (service *metricsService) SaveAll(ctx context.Context, metrics []models.Metrics) *api.APIError {
	counterSums := make(map[string]int64)
	gaugeLastValues := make(map[string]float64)

	for _, metric := range metrics {
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
		go func() { errChan <- service.repository.AddAll(ctx, counters) }()
	} else {
		go func() { errChan <- nil }()
	}

	if len(gauges) > 0 {
		go func() { errChan <- service.repository.ResetAll(ctx, gauges) }()
	} else {
		go func() { errChan <- nil }()
	}

	err1 := <-errChan
	err2 := <-errChan

	if err1 != nil || err2 != nil {
		return api.Internal("save metrics error", fmt.Errorf("counters: %v, gauges: %v", err1, err2))
	}

	go service.notifyMany(ctx, metrics)

	return nil
}
