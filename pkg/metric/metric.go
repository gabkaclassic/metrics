package metric

type MetricType string

const (
	GaugeType   MetricType = "gauge"
	CounterType MetricType = "counter"
)

type Metric interface {
	Type() MetricType
	Name() string
	Update()
	Value() any
}
