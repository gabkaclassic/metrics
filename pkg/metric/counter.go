package metric

type Counter struct {
	Metric
	value int64
}

func (metric Counter) Type() MetricType {
	return CounterType
}

func (metric *Counter) Value() any {

	return metric.value
}

type PollCount struct {
	Counter
}

func (metric *PollCount) Update() {
	metric.value += 1
}

func (metric PollCount) Name() string {
	return "pollCount"
}
