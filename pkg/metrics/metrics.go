package metrics

type MetricType string

const (
	GaugeType   MetricType = "gauge"
	CounterType MetricType = "counter"
)

type Metric[T float64 | int64] interface {
	Type() MetricType
	Name() string
	Value() T
}
