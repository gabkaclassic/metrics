// Package grpcserver provides gRPC server implementation for metrics processing.
//
// Architecture overview:
//
// The package contains:
//   - MetricsGRPCServer — business gRPC handler implementation.
//   - GRPCServer — transport server wrapper with lifecycle management.
//   - Unary interceptors for:
//   - audit context propagation
//   - IP address trust validation
//   - request logging
//
// Server lifecycle:
//
// Server is created via SetupGRPCServer, which:
//
//   - Registers Metrics service implementation.
//   - Configures interceptor chain.
//   - Validates trusted CIDR configuration.
//
// Runtime behaviour:
//
// The server supports graceful shutdown using context cancellation.
// When shutdown is requested:
//   - Active RPC calls are completed.
//   - gRPC server stops accepting new connections.
//   - Server waits until Serve() loop terminates.
//
// Security:
//
// IP restriction can be configured via CIDR whitelist.
// Incoming requests are validated against X-Real-IP metadata header.
//
// Context propagation:
//
// The server propagates:
//   - Client source IP
//   - Request timestamp
//
// Error handling:
//
// Business-layer API errors are converted to gRPC status codes.
package grpcserver
