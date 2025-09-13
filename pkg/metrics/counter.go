package metrics

type Counter struct {
	Metric[int64]
}

func (metric *Counter) Type() MetricType {
	return CounterType
}

type PollCount struct {
	Counter
}

func (metric *PollCount) Value() int64 {
	return 1
}

func (metric *PollCount) Name() string {
	return "pollCount"
}
