package metrics

import (
	"math/rand"
	"runtime"
)

type Gauge struct {
	Metric[float64]
}

func (metric *Gauge) Type() MetricType {
	return GaugeType
}

type RandomValue struct {
	Gauge
}

func (metric *RandomValue) Value() float64 {
	return rand.Float64()
}

func (metric *RandomValue) Name() string {
	return "randomValue"
}

type RuntimeGauge struct {
	Gauge
	stats *runtime.MemStats
}

func (gauge *RuntimeGauge) Stats() *runtime.MemStats {
	return gauge.stats
}

func (gauge *RuntimeGauge) SetStats(stats *runtime.MemStats) {
	gauge.stats = stats
}

type Alloc struct {
	RuntimeGauge
}

func (metric *Alloc) Value() float64 {
	return float64(metric.Stats().Alloc)
}

func (metric *Alloc) Name() string {
	return "alloc"
}

type BuckHashSys struct {
	RuntimeGauge
}

func (metric *BuckHashSys) Value() float64 {
	return float64(metric.Stats().BuckHashSys)
}

func (metric *BuckHashSys) Name() string {
	return "buckHashSys"
}

type Frees struct {
	RuntimeGauge
}

func (metric *Frees) Value() float64 {
	return float64(metric.Stats().Frees)
}

func (metric *Frees) Name() string {
	return "frees"
}

type GCCPUFraction struct {
	RuntimeGauge
}

func (metric *GCCPUFraction) Value() float64 {
	return float64(metric.Stats().GCCPUFraction)
}

func (metric *GCCPUFraction) Name() string {
	return "GCCPUFraction"
}

type GCSys struct {
	RuntimeGauge
}

func (metric *GCSys) Value() float64 {
	return float64(metric.Stats().GCSys)
}

func (metric *GCSys) Name() string {
	return "GCSys"
}

type HeapAlloc struct {
	RuntimeGauge
}

func (metric *HeapAlloc) Value() float64 {
	return float64(metric.Stats().HeapAlloc)
}

func (metric *HeapAlloc) Name() string {
	return "heapAlloc"
}

type HeapIdle struct {
	RuntimeGauge
}

func (metric *HeapIdle) Value() float64 {
	return float64(metric.Stats().HeapIdle)
}

func (metric *HeapIdle) Name() string {
	return "HeapIdle"
}

type HeapInuse struct {
	RuntimeGauge
}

func (metric *HeapInuse) Value() float64 {
	return float64(metric.Stats().HeapInuse)
}

func (metric *HeapInuse) Name() string {
	return "HeapInuse"
}

type HeapObjects struct {
	RuntimeGauge
}

func (metric *HeapObjects) Value() float64 {
	return float64(metric.Stats().HeapObjects)
}

func (metric *HeapObjects) Name() string {
	return "HeapObjects"
}

type HeapReleased struct {
	RuntimeGauge
}

func (metric *HeapReleased) Value() float64 {
	return float64(metric.Stats().HeapReleased)
}

func (metric *HeapReleased) Name() string {
	return "HeapReleased"
}

type HeapSys struct {
	RuntimeGauge
}

func (metric *HeapSys) Value() float64 {
	return float64(metric.Stats().HeapSys)
}

func (metric *HeapSys) Name() string {
	return "HeapSys"
}

type LastGC struct {
	RuntimeGauge
}

func (metric *LastGC) Value() float64 {
	return float64(metric.Stats().LastGC)
}

func (metric *LastGC) Name() string {
	return "LastGC"
}

type Lookups struct {
	RuntimeGauge
}

func (metric *Lookups) Value() float64 {
	return float64(metric.Stats().Lookups)
}

func (metric *Lookups) Name() string {
	return "Lookups"
}

type MCacheInuse struct {
	RuntimeGauge
}

func (metric *MCacheInuse) Value() float64 {
	return float64(metric.Stats().MCacheInuse)
}

func (metric *MCacheInuse) Name() string {
	return "MCacheInuse"
}

type MCacheSys struct {
	RuntimeGauge
}

func (metric *MCacheSys) Value() float64 {
	return float64(metric.Stats().MCacheSys)
}

func (metric *MCacheSys) Name() string {
	return "MCacheSys"
}

type MSpanInuse struct {
	RuntimeGauge
}

func (metric *MSpanInuse) Value() float64 {
	return float64(metric.Stats().MSpanInuse)
}

func (metric *MSpanInuse) Name() string {
	return "MSpanInuse"
}

type MSpanSys struct {
	RuntimeGauge
}

func (metric *MSpanSys) Value() float64 {
	return float64(metric.Stats().MSpanSys)
}

func (metric *MSpanSys) Name() string {
	return "MSpanSys"
}

type Mallocs struct {
	RuntimeGauge
}

func (metric *Mallocs) Value() float64 {
	return float64(metric.Stats().Mallocs)
}

func (metric *Mallocs) Name() string {
	return "Mallocs"
}

type NextGC struct {
	RuntimeGauge
}

func (metric *NextGC) Value() float64 {
	return float64(metric.Stats().NextGC)
}

func (metric *NextGC) Name() string {
	return "NextGC"
}

type NumForcedGC struct {
	RuntimeGauge
}

func (metric *NumForcedGC) Value() float64 {
	return float64(metric.Stats().NumForcedGC)
}

func (metric *NumForcedGC) Name() string {
	return "NumForcedGC"
}

type NumGC struct {
	RuntimeGauge
}

func (metric *NumGC) Value() float64 {
	return float64(metric.Stats().NumGC)
}

func (metric *NumGC) Name() string {
	return "NumGC"
}

type OtherSys struct {
	RuntimeGauge
}

func (metric *OtherSys) Value() float64 {
	return float64(metric.Stats().OtherSys)
}

func (metric *OtherSys) Name() string {
	return "OtherSys"
}

type PauseTotalNs struct {
	RuntimeGauge
}

func (metric *PauseTotalNs) Value() float64 {
	return float64(metric.Stats().PauseTotalNs)
}

func (metric *PauseTotalNs) Name() string {
	return "PauseTotalNs"
}

type StackInuse struct {
	RuntimeGauge
}

func (metric *StackInuse) Value() float64 {
	return float64(metric.Stats().StackInuse)
}

func (metric *StackInuse) Name() string {
	return "StackInuse"
}

type StackSys struct {
	RuntimeGauge
}

func (metric *StackSys) Value() float64 {
	return float64(metric.Stats().StackSys)
}

func (metric *StackSys) Name() string {
	return "StackSys"
}

type Sys struct {
	RuntimeGauge
}

func (metric *Sys) Value() float64 {
	return float64(metric.Stats().Sys)
}

func (metric *Sys) Name() string {
	return "Sys"
}

type TotalAlloc struct {
	RuntimeGauge
}

func (metric *TotalAlloc) Value() float64 {
	return float64(metric.Stats().TotalAlloc)
}

func (metric *TotalAlloc) Name() string {
	return "TotalAlloc"
}
