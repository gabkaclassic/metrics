// Package repository provides data access layer implementations for metrics storage.
//
// The repository pattern abstracts storage details and provides:
//   - Business logic for metric operations
//   - Thread-safe concurrent access
//   - Consistent error handling
//   - Support for different storage backends
package repository
