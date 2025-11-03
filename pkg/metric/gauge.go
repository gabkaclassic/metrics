package metric

import (
	"github.com/shirou/gopsutil/v4/mem"
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
	return "RandomValue"
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

func PsMetrics(getMem func() *mem.VirtualMemoryStat, getCPU func() *[]float64) []Metric {
	return []Metric{
		NewRuntimeGaugeMetric("TotalMemory", func() float64 {
			m := getMem()
			if m == nil {
				return 0
			}
			return float64(m.Total)
		}),
		NewRuntimeGaugeMetric("FreeMemory", func() float64 {
			m := getMem()
			if m == nil {
				return 0
			}
			return float64(m.Free)
		}),
		NewRuntimeGaugeMetric("CPUutilization1", func() float64 {
			cpu := getCPU()
			if cpu == nil || len(*cpu) == 0 {
				return 0
			}
			return (*cpu)[0]
		}),
	}
}

func RuntimeMetrics(stats *runtime.MemStats) []Metric {
	return []Metric{
		NewRuntimeGaugeMetric(
			"Alloc", func() float64 { return float64(stats.Alloc) },
		),
		NewRuntimeGaugeMetric(
			"BuckHashSys", func() float64 { return float64(stats.BuckHashSys) },
		),
		NewRuntimeGaugeMetric(
			"Frees", func() float64 { return float64(stats.Frees) },
		),
		NewRuntimeGaugeMetric(
			"GCCPUFraction", func() float64 { return float64(stats.GCCPUFraction) },
		),
		NewRuntimeGaugeMetric(
			"GCSys", func() float64 { return float64(stats.Alloc) },
		),
		NewRuntimeGaugeMetric(
			"HeapAlloc", func() float64 { return float64(stats.HeapAlloc) },
		),
		NewRuntimeGaugeMetric(
			"HeapIdle", func() float64 { return float64(stats.HeapIdle) },
		),
		NewRuntimeGaugeMetric(
			"HeapInuse", func() float64 { return float64(stats.HeapInuse) },
		),
		NewRuntimeGaugeMetric(
			"HeapObjects", func() float64 { return float64(stats.HeapObjects) },
		),
		NewRuntimeGaugeMetric(
			"HeapReleased", func() float64 { return float64(stats.HeapReleased) },
		),
		NewRuntimeGaugeMetric(
			"HeapSys", func() float64 { return float64(stats.HeapSys) },
		),
		NewRuntimeGaugeMetric(
			"LastGC", func() float64 { return float64(stats.LastGC) },
		),
		NewRuntimeGaugeMetric(
			"Lookups", func() float64 { return float64(stats.Lookups) },
		),
		NewRuntimeGaugeMetric(
			"MCacheInuse", func() float64 { return float64(stats.MCacheInuse) },
		),
		NewRuntimeGaugeMetric(
			"MCacheSys", func() float64 { return float64(stats.MCacheSys) },
		),
		NewRuntimeGaugeMetric(
			"MCacheInuse", func() float64 { return float64(stats.MCacheInuse) },
		),
		NewRuntimeGaugeMetric(
			"MSpanInuse", func() float64 { return float64(stats.MSpanInuse) },
		),
		NewRuntimeGaugeMetric(
			"MSpanSys", func() float64 { return float64(stats.MSpanSys) },
		),
		NewRuntimeGaugeMetric(
			"Mallocs", func() float64 { return float64(stats.Mallocs) },
		),
		NewRuntimeGaugeMetric(
			"NextGC", func() float64 { return float64(stats.NextGC) },
		),
		NewRuntimeGaugeMetric(
			"NumForcedGC", func() float64 { return float64(stats.NumForcedGC) },
		),
		NewRuntimeGaugeMetric(
			"NumGC", func() float64 { return float64(stats.NumGC) },
		),
		NewRuntimeGaugeMetric(
			"NumForcedGC", func() float64 { return float64(stats.NumForcedGC) },
		),
		NewRuntimeGaugeMetric(
			"OtherSys", func() float64 { return float64(stats.OtherSys) },
		),
		NewRuntimeGaugeMetric(
			"PauseTotalNs", func() float64 { return float64(stats.PauseTotalNs) },
		),
		NewRuntimeGaugeMetric(
			"NumForcedGC", func() float64 { return float64(stats.NumForcedGC) },
		),
		NewRuntimeGaugeMetric(
			"StackInuse", func() float64 { return float64(stats.StackInuse) },
		),
		NewRuntimeGaugeMetric(
			"StackSys", func() float64 { return float64(stats.StackSys) },
		),
		NewRuntimeGaugeMetric(
			"Sys", func() float64 { return float64(stats.Sys) },
		),
		NewRuntimeGaugeMetric(
			"StackInuse", func() float64 { return float64(stats.StackInuse) },
		),
		NewRuntimeGaugeMetric(
			"TotalAlloc", func() float64 { return float64(stats.TotalAlloc) },
		),
	}
}
