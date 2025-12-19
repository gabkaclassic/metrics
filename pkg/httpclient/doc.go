// Package httpclient provides a configurable HTTP client with retry,
// delay, and response filtering support.
//
// Features:
//   - Base URL configuration
//   - Custom headers and query parameters
//   - Automatic retries with configurable delay
//   - Response filtering to decide retry logic
//   - Timeout per request
//
// The package exposes an interface HTTPClient and a concrete Client
// with functional options for flexible configuration.
package httpclient
