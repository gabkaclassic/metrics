package metric

import (
	"math/rand"
	"runtime"

	"github.com/shirou/gopsutil/v4/mem"
)

// ValueFunctionType is a function that returns a gauge value.
type ValueFunctionType func() float64

// Gauge is the interface for gauge-type metrics.
type Gauge interface {
	Metric
}

// GaugeMetric provides the base implementation for gauge metrics.
type GaugeMetric struct {
	value  float64
	_value ValueFunctionType
}

// Type returns GaugeType for GaugeMetric instances.
func (metric GaugeMetric) Type() MetricType {
	return GaugeType
}

// Value returns the current gauge value.
func (metric *GaugeMetric) Value() any {
	return metric.value
}

// RandomValue is a gauge that generates random float64 values.
type RandomValue struct {
	GaugeMetric
}

// Name returns "RandomValue" as the metric identifier.
func (metric *RandomValue) Name() string {
	return "RandomValue"
}

// Update generates a new random value between 0.0 and 1.0.
func (metric *RandomValue) Update() {
	metric.value = rand.Float64()
}

// RuntimeGaugeMetric is a gauge that computes its value using a function.
type RuntimeGaugeMetric struct {
	GaugeMetric
	name string
}

// Update executes the value function to refresh the metric.
func (metric *RuntimeGaugeMetric) Update() {
	metric.value = metric._value()
}

// Name returns the runtime gauge's name.
func (metric *RuntimeGaugeMetric) Name() string {
	return metric.name
}

// NewRuntimeGaugeMetric creates a new RuntimeGaugeMetric.
//
// name: The metric identifier
// value: Function that returns the metric value when called
func NewRuntimeGaugeMetric(name string, value ValueFunctionType) *RuntimeGaugeMetric {
	return &RuntimeGaugeMetric{
		GaugeMetric: GaugeMetric{
			_value: value,
		},
		name: name,
	}
}

// PsMetrics returns system metrics from procps (memory and CPU).
//
// getMem: Function that returns VirtualMemoryStat
// getCPU: Function that returns CPU utilization percentages
//
// Returns a slice containing:
//   - TotalMemory: Total system memory in bytes
//   - FreeMemory: Available system memory in bytes
//   - CPUutilization1: Current CPU utilization percentage
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

// RuntimeMetrics returns Go runtime memory statistics.
//
// stats: Pointer to runtime.MemStats struct containing current runtime metrics
//
// Returns a slice of 28 runtime metrics including:
//   - Alloc: Currently allocated heap objects bytes
//   - HeapAlloc: Heap allocation bytes
//   - NumGC: Number of completed GC cycles
//   - GCCPUFraction: CPU time used by GC
//   - And other runtime statistics
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
