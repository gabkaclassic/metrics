// Package models provides data structures for metrics transmission and storage.
package models

const (
	Counter = "counter"
	Gauge   = "gauge"
)

// Metrics represents a metric data structure for JSON serialization.
// It's used for communication between agent and server.
type Metrics struct {
	ID    string   `json:"id"`              // Metric name identifier
	MType string   `json:"type"`            // Metric type: "gauge" or "counter"
	Delta *int64   `json:"delta,omitempty"` // Counter metric value (increment)
	Value *float64 `json:"value,omitempty"` // Gauge metric value (absolute)
	Hash  string   `json:"hash,omitempty"`
}
