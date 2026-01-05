// Package reset provides utilities for working with resettable values.
//
// The package defines the Resetable interface and a generic ResetablePool,
// which stores values implementing Resetable and guarantees that each value
// is reset before being reused.
//
// ResetablePool is safe for concurrent use.
package reset

import "sync"

// Resetable represents a value that can reset its internal state
// to a well-defined zero or initial state.
type Resetable interface {
	Reset()
}

// ResetablePool is a simple concurrent-safe pool for values implementing
// Resetable.
//
// Values returned to the pool via Put are reset immediately.
// Values retrieved via Get are returned by value, not by pointer.
type ResetablePool[T Resetable] struct {
	values []T
	mu     sync.Mutex
}

// New creates a new empty ResetablePool.
func New[T Resetable]() *ResetablePool[T] {
	return &ResetablePool[T]{
		values: make([]T, 0),
	}
}

// Get returns a value from the pool.
//
// If the pool is empty, the zero value of T and false are returned.
// Otherwise, a value and true are returned.
func (p *ResetablePool[T]) Get() (T, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.values) == 0 {
		var zero T
		return zero, false
	}

	value := p.values[0]
	p.values = p.values[1:]
	return value, true
}

// Put resets the value and returns it to the pool.
func (p *ResetablePool[T]) Put(value T) {
	p.mu.Lock()
	defer p.mu.Unlock()

	value.Reset()
	p.values = append(p.values, value)
}
