package metric

// Counter represents an integer counter metric.
type Counter struct {
	value int64
}

// Type returns CounterType for Counter metrics.
func (metric Counter) Type() MetricType {
	return CounterType
}

// Value returns the current counter value.
func (metric *Counter) Value() any {
	return metric.value
}

// PollCount is a specific counter that tracks the number of polling cycles.
type PollCount struct {
	Counter
}

// Update increments the PollCount by 1.
func (metric *PollCount) Update() {
	metric.value += 1
}

// Name returns "PollCount" as the metric identifier.
func (metric PollCount) Name() string {
	return "PollCount"
}
