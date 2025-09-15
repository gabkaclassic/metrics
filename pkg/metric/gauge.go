package metric

import (
	"math/rand"
	"runtime"
)

type ValueFunctionType func() float64

type Gauge interface {
	Metric
}

type GaugeMetric struct {
	Gauge
	value  float64
	_value ValueFunctionType
}

func (metric GaugeMetric) Type() MetricType {
	return GaugeType
}

func (metric *GaugeMetric) Value() any {
	return metric.value
}

type RandomValue struct {
	GaugeMetric
}

func (metric *RandomValue) Name() string {
	return "randomValue"
}

func (metric *RandomValue) Update() {
	metric.value = rand.Float64()
}

type RuntimeGaugeMetric struct {
	GaugeMetric
	name string
}

func (metric *RuntimeGaugeMetric) Update() {
	metric.value = metric._value()
}

func (metric *RuntimeGaugeMetric) Name() string {
	return metric.name
}

func NewRuntimeGaugeMetric(name string, value ValueFunctionType) *RuntimeGaugeMetric {
	return &RuntimeGaugeMetric{
		GaugeMetric: GaugeMetric{
			_value: value,
		},
		name: name,
	}
}

func RuntimeMetrics(stats *runtime.MemStats) []Metric {
	return []Metric{
		NewRuntimeGaugeMetric(
			"alloc", func() float64 { return float64(stats.Alloc) },
		),
		NewRuntimeGaugeMetric(
			"buckHashSys", func() float64 { return float64(stats.BuckHashSys) },
		),
		NewRuntimeGaugeMetric(
			"frees", func() float64 { return float64(stats.Frees) },
		),
		NewRuntimeGaugeMetric(
			"GCCPUFraction", func() float64 { return float64(stats.GCCPUFraction) },
		),
		NewRuntimeGaugeMetric(
			"GCSys", func() float64 { return float64(stats.Alloc) },
		),
		NewRuntimeGaugeMetric(
			"heapAlloc", func() float64 { return float64(stats.HeapAlloc) },
		),
		NewRuntimeGaugeMetric(
			"heapIdle", func() float64 { return float64(stats.HeapIdle) },
		),
		NewRuntimeGaugeMetric(
			"heapInuse", func() float64 { return float64(stats.HeapInuse) },
		),
		NewRuntimeGaugeMetric(
			"heapObjects", func() float64 { return float64(stats.HeapObjects) },
		),
		NewRuntimeGaugeMetric(
			"heapReleased", func() float64 { return float64(stats.HeapReleased) },
		),
		NewRuntimeGaugeMetric(
			"heapSys", func() float64 { return float64(stats.HeapSys) },
		),
		NewRuntimeGaugeMetric(
			"lastGC", func() float64 { return float64(stats.LastGC) },
		),
		NewRuntimeGaugeMetric(
			"lookups", func() float64 { return float64(stats.Lookups) },
		),
		NewRuntimeGaugeMetric(
			"mCacheInuse", func() float64 { return float64(stats.MCacheInuse) },
		),
		NewRuntimeGaugeMetric(
			"mCacheSys", func() float64 { return float64(stats.MCacheSys) },
		),
		NewRuntimeGaugeMetric(
			"mCacheInuse", func() float64 { return float64(stats.MCacheInuse) },
		),
		NewRuntimeGaugeMetric(
			"mSpanInuse", func() float64 { return float64(stats.MSpanInuse) },
		),
		NewRuntimeGaugeMetric(
			"mSpanSys", func() float64 { return float64(stats.MSpanSys) },
		),
		NewRuntimeGaugeMetric(
			"mallocs", func() float64 { return float64(stats.Mallocs) },
		),
		NewRuntimeGaugeMetric(
			"nextGC", func() float64 { return float64(stats.NextGC) },
		),
		NewRuntimeGaugeMetric(
			"numForcedGC", func() float64 { return float64(stats.NumForcedGC) },
		),
		NewRuntimeGaugeMetric(
			"numGC", func() float64 { return float64(stats.NumGC) },
		),
		NewRuntimeGaugeMetric(
			"numForcedGC", func() float64 { return float64(stats.NumForcedGC) },
		),
		NewRuntimeGaugeMetric(
			"otherSys", func() float64 { return float64(stats.OtherSys) },
		),
		NewRuntimeGaugeMetric(
			"pauseTotalNs", func() float64 { return float64(stats.PauseTotalNs) },
		),
		NewRuntimeGaugeMetric(
			"numForcedGC", func() float64 { return float64(stats.NumForcedGC) },
		),
		NewRuntimeGaugeMetric(
			"stackInuse", func() float64 { return float64(stats.StackInuse) },
		),
		NewRuntimeGaugeMetric(
			"stackSys", func() float64 { return float64(stats.StackSys) },
		),
		NewRuntimeGaugeMetric(
			"sys", func() float64 { return float64(stats.Sys) },
		),
		NewRuntimeGaugeMetric(
			"stackInuse", func() float64 { return float64(stats.StackInuse) },
		),
		NewRuntimeGaugeMetric(
			"totalAlloc", func() float64 { return float64(stats.TotalAlloc) },
		),
	}
}
