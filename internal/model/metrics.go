// Package models provides data structures for metrics transmission and storage.
package models

const (
	Counter = "counter"
	Gauge   = "gauge"
)

// Metrics represents a metric entity exchanged between agent and server.
//
// Supports two metric types:
//   - gauge   — absolute floating-point value
//   - counter — incremental integer value
//
// Exactly one of `value` or `delta` must be set depending on metric type.
//
// swagger:model Metrics
type Metrics struct {
	// Metric identifier (name).
	// required: true
	ID string `json:"id"`

	// Metric type.
	// required: true
	// enum: gauge,counter
	MType string `json:"type"`

	// Counter increment value.
	// Used only when type is "counter".
	// example: 42
	Delta *int64 `json:"delta,omitempty"`

	// Gauge absolute value.
	// Used only when type is "gauge".
	// example: 3.14
	Value *float64 `json:"value,omitempty"`

	// Optional integrity hash.
	// example: 1a2b3c4d
	Hash string `json:"hash,omitempty"`
}
