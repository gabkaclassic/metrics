// Package metric provides the core metric abstraction and implementations.
//
// The package defines two metric types:
//   - Gauge: Represents a value that can go up and down (e.g., memory usage)
//   - Counter: Represents a monotonically increasing value (e.g., request count)
package metric

// MetricType defines the type of a metric.
type MetricType string

const (
	GaugeType   MetricType = "gauge"   // Gauge metric type
	CounterType MetricType = "counter" // Counter metric type
)

// Metric is the interface that all metrics must implement.
type Metric interface {
	// Type returns the metric type (gauge or counter).
	Type() MetricType

	// Name returns the unique identifier of the metric.
	Name() string

	// Update refreshes the metric value (e.g., reads from system).
	Update()

	// Value returns the current metric value as interface{}.
	Value() any
}
