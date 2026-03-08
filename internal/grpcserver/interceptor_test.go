package grpcserver

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestAuditUnaryInterceptor(t *testing.T) {
	tests := []struct {
		name     string
		metaIP   string
		expectIP string
	}{
		{
			name:     "x-real-ip header present",
			metaIP:   "1.2.3.4",
			expectIP: "1.2.3.4",
		},
		{
			name:     "x-real-ip header with spaces",
			metaIP:   "  5.6.7.8  ",
			expectIP: "5.6.7.8",
		},
		{
			name:     "missing x-real-ip header",
			metaIP:   "",
			expectIP: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedIP string
			var capturedTS int64

			handler := func(ctx context.Context, req any) (any, error) {
				ip, _ := ctx.Value(ctxSoureIPKey).(string)
				ts, _ := ctx.Value(ctxTSKey).(int64)

				capturedIP = ip
				capturedTS = ts

				return nil, nil
			}

			interceptor := auditUnaryInterceptor()

			ctx := context.Background()

			if tt.metaIP != "" {
				md := metadata.Pairs("x-real-ip", tt.metaIP)
				ctx = metadata.NewIncomingContext(ctx, md)
			}

			info := &grpc.UnaryServerInfo{
				FullMethod: "/test.Service/Test",
			}

			_, err := interceptor(ctx, nil, info, handler)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectIP, capturedIP)
			assert.Greater(t, capturedTS, int64(0))
		})
	}
}

func TestTrustAddressUnaryInterceptor(t *testing.T) {
	tests := []struct {
		name        string
		cidr        string
		metaIP      string
		expectError bool
		expectCall  bool
	}{
		{
			name:        "empty cidr allows request",
			cidr:        "",
			metaIP:      "1.2.3.4",
			expectError: false,
			expectCall:  true,
		},
		{
			name:        "ip inside trusted cidr passes",
			cidr:        "1.2.3.0/24",
			metaIP:      "1.2.3.4",
			expectError: false,
			expectCall:  true,
		},
		{
			name:        "ip outside trusted cidr rejected",
			cidr:        "1.2.3.0/24",
			metaIP:      "5.6.7.8",
			expectError: true,
			expectCall:  false,
		},
		{
			name:        "invalid ip rejected",
			cidr:        "1.2.3.0/24",
			metaIP:      "invalid-ip",
			expectError: true,
			expectCall:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			nextCalled := false

			handler := func(ctx context.Context, req any) (any, error) {
				nextCalled = true
				return nil, nil
			}

			interceptor, err := trustAddressUnaryInterceptor(tt.cidr)
			assert.NoError(t, err)

			ctx := context.Background()

			if tt.metaIP != "" {
				md := metadata.Pairs("x-real-ip", tt.metaIP)
				ctx = metadata.NewIncomingContext(ctx, md)
			}

			info := &grpc.UnaryServerInfo{
				FullMethod: "/test.Service/Test",
			}

			resp, err := interceptor(ctx, nil, info, handler)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectCall, nextCalled)
		})
	}
}
