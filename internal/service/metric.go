// Package service provides business logic layer for metrics operations.
//
// The service layer coordinates between HTTP handlers and data repositories,
// providing:
//   - Business logic validation and processing
//   - Error handling and API error formatting
//   - Audit logging for security and compliance
//   - Concurrent batch processing
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

// MetricsService defines the interface for metric business operations.
// Handles validation, processing, and coordinates repository and audit systems.

type MetricsService interface {
	// Get retrieves a metric value by ID and type.
	// Returns the raw value (int64 for counters, float64 for gauges).
	Get(context.Context, string, string) (any, *api.APIError)

	// GetStruct retrieves a complete metric structure by ID and type.
	// Returns the full Metrics model including all fields.
	GetStruct(context.Context, string, string) (models.Metrics, *api.APIError)

	// Save processes and stores a metric from raw string values.
	// Parses and validates input before storage.
	Save(context.Context, string, string, string) *api.APIError

	// SaveStruct stores a pre-validated metric structure.
	// Used for JSON API endpoints with structured input.
	SaveStruct(context.Context, models.Metrics) *api.APIError

	// SaveAll processes and stores multiple metrics efficiently.
	// Aggregates counters and processes gauges concurrently.
	SaveAll(context.Context, []models.Metrics) *api.APIError

	// GetAll retrieves all stored metrics as a map.
	// Returns metric ID to value mapping (int64 or float64).
	GetAll(context.Context) (map[string]any, *api.APIError)
}

// metricsService implements MetricsService with repository and audit integration.
// Provides thread-safe operations through repository synchronization.

type metricsService struct {
	repository repository.MetricsRepository
	auditor    audit.Auditor
}

// NewMetricsService creates a new metrics service with required dependencies.
//
// repository: Data access layer for metric storage operations
// auditor: Audit logging system for security and compliance tracking
//
// Returns:
//   - MetricsService: Ready-to-use service instance
//   - error: If repository or auditor is nil
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

// notifyOne logs a single metric operation to the audit system.
// Extracts timestamp and IP from context and records asynchronously.
func (service *metricsService) notifyOne(ctx context.Context, metric models.Metrics) {
	ts := middleware.AuditTSFromCtx(ctx)

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

// notifyMany logs multiple metric operations to the audit system.
// Extracts timestamp and IP from context and records asynchronously.
func (service *metricsService) notifyMany(ctx context.Context, metrics []models.Metrics) {
	if len(metrics) == 0 {
		slog.Debug("Metrics list for audit is empty")
		return
	}

	ts := middleware.AuditTSFromCtx(ctx)

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

// GetAll retrieves all metrics from the repository.
// Returns API error if repository operation fails.
func (service *metricsService) GetAll(ctx context.Context) (map[string]any, *api.APIError) {
	metrics, err := service.repository.GetAll(ctx)

	if err != nil {
		return nil, api.Internal("Get all metrics error", err)
	}

	return metrics, nil
}

// Get retrieves a single metric value by ID and type.
// Validates that the retrieved metric matches the requested type.
// Returns the appropriate value based on metric type.
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

// GetStruct retrieves a complete metric structure by ID and type.
// Returns the full metric model with all fields populated.
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

// Save processes and stores a metric from raw string inputs.
// Validates metric type, parses value, and calls appropriate repository method.
// Performs audit logging asynchronously after successful storage.
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
			err := service.repository.ResetOne(ctx, metric)
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

// SaveStruct stores a pre-validated metric structure.
// Routes to appropriate repository method based on metric type.
// Performs audit logging asynchronously after successful storage.
func (service *metricsService) SaveStruct(ctx context.Context, metric models.Metrics) *api.APIError {
	var err error
	switch metric.MType {
	case models.Counter:
		err = service.repository.Add(ctx, metric)
	case models.Gauge:
		err = service.repository.ResetOne(ctx, metric)
	default:
		return api.BadRequest(fmt.Sprintf("invalid metric type: %s", metric.MType))
	}

	if err != nil {
		return api.Internal("save metric error", err)
	}
	go service.notifyOne(ctx, metric)

	return nil
}

// SaveAll efficiently processes and stores multiple metrics.
// Aggregates counter deltas and processes counters/gauges concurrently.
// Performs audit logging asynchronously for all metrics.
//
// Process:
//  1. Aggregates counter deltas by metric ID
//  2. Collects latest gauge values by metric ID
//  3. Processes counters and gauges in parallel goroutines
//  4. Returns combined error if any operation fails
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
