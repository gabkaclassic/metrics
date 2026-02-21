package grpcserver

import (
	"context"
	"log/slog"
	"net"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// auditUnaryInterceptor adds audit context information into request context.
//
// Behaviour:
//
// Extracts client IP address from incoming gRPC metadata.
// Priority of IP detection:
//  1. "x-real-ip" metadata header.
//  2. Default value "unknown" if header is missing.
//
// Context values injected:
//
//   - key "sourceIP" → client IP address (string)
//   - key "ts" → Unix timestamp (int64) representing request start time.
//
// Purpose:
//
// Enables downstream handlers and services to access audit information.
func auditUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {

		ip := "unknown"

		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if vals := md.Get("x-real-ip"); len(vals) > 0 {
				ip = strings.TrimSpace(vals[0])
			}
		}

		ctx = context.WithValue(ctx, "sourceIP", ip)
		ctx = context.WithValue(ctx, "ts", time.Now().Unix())

		return handler(ctx, req)
	}
}

// trustAddressUnaryInterceptor creates IP whitelist validation interceptor.
//
// Parameters:
//
// cidr: CIDR network string used as whitelist filter.
//
// Behaviour:
//
// If cidr is empty:
//   - Returns passthrough interceptor without validation.
//
// If cidr is provided:
//   - Parses CIDR network.
//   - Extracts client IP from "x-real-ip" metadata header.
//   - Rejects request if:
//   - IP is invalid, or
//   - IP is outside trusted network.
//
// Errors:
//
// Returns error during interceptor creation if CIDR parsing fails.
//
// Runtime error response:
//
// Uses gRPC PermissionDenied status code for unauthorized requests.
func trustAddressUnaryInterceptor(cidr string) (grpc.UnaryServerInterceptor, error) {
	if len(cidr) == 0 {
		return func(
			ctx context.Context,
			req any,
			info *grpc.UnaryServerInfo,
			handler grpc.UnaryHandler,
		) (any, error) {
			return handler(ctx, req)
		}, nil
	}

	_, trustedCIDR, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {

		ipStr := ""

		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if vals := md.Get("x-real-ip"); len(vals) > 0 {
				ipStr = strings.TrimSpace(vals[0])
			}
		}

		ip := net.ParseIP(ipStr)

		if ip == nil || !trustedCIDR.Contains(ip) {
			slog.Info("Forbidden by untrusted IP",
				slog.String("IP", ipStr),
			)
			return nil, status.Error(codes.PermissionDenied, "forbidden")
		}

		return handler(ctx, req)
	}, nil
}

// loggingUnaryInterceptor logs gRPC request execution metrics.
//
// Logged fields:
//
//   - method: Full RPC method name.
//   - duration: Request processing time.
//
// Behaviour:
//
// Measures time from request entry to handler completion.
// Logs request execution using structured logging.
//
// Purpose:
//
// Provides observability for RPC performance monitoring.
func loggingUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {

		start := time.Now()

		resp, err := handler(ctx, req)

		duration := time.Since(start)

		slog.Info("gRPC request",
			slog.String("method", info.FullMethod),
			slog.Duration("duration", duration),
		)

		return resp, err
	}
}
