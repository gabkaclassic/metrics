package grpcserver

import (
	"context"
	"net/http"
	"testing"

	models "github.com/gabkaclassic/metrics/internal/model"
	pb "github.com/gabkaclassic/metrics/internal/proto"
	"github.com/gabkaclassic/metrics/internal/service"
	api "github.com/gabkaclassic/metrics/pkg/error"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func floatPtr(f float64) *float64 { return &f }
func intPtr(i int64) *int64       { return &i }

func TestMetricsGRPCServer_UpdateMetrics(t *testing.T) {
	tests := []struct {
		name           string
		req            *pb.UpdateMetricsRequest
		mockSetup      func(mockSvc *service.MockMetricsService)
		expectError    bool
		expectErrorMsg string
		expectResponse bool
	}{
		{
			name: "nil request",
			req:  nil,
			mockSetup: func(mockSvc *service.MockMetricsService) {
			},
			expectError:    false,
			expectResponse: true,
		},
		{
			name: "empty metrics batch",
			req:  &pb.UpdateMetricsRequest{Metrics: []*pb.Metric{}},
			mockSetup: func(mockSvc *service.MockMetricsService) {
			},
			expectError:    false,
			expectResponse: true,
		},
		{
			name: "valid metrics batch",
			req: &pb.UpdateMetricsRequest{
				Metrics: []*pb.Metric{
					{
						Id:    "m1",
						Type:  pb.Metric_COUNTER,
						Delta: 10,
					},
				},
			},
			mockSetup: func(mockSvc *service.MockMetricsService) {
				mockSvc.EXPECT().
					SaveAll(mock.Anything, mock.Anything).
					Return(nil)
			},
			expectError:    false,
			expectResponse: true,
		},
		{
			name: "service layer error",
			req: &pb.UpdateMetricsRequest{
				Metrics: []*pb.Metric{
					{
						Id:    "m2",
						Type:  pb.Metric_GAUGE,
						Value: 1.5,
					},
				},
			},
			mockSetup: func(mockSvc *service.MockMetricsService) {
				mockSvc.EXPECT().
					SaveAll(mock.Anything, mock.Anything).
					Return(api.BadRequest("save failed"))
			},
			expectError:    true,
			expectErrorMsg: "save failed",
			expectResponse: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockSvc := service.NewMockMetricsService(t)
			tt.mockSetup(mockSvc)

			server := NewMetricsGRPCServer(mockSvc)

			ctx := context.Background()

			resp, err := server.UpdateMetrics(ctx, tt.req)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectErrorMsg)
				assert.Nil(t, resp)
				return
			}

			assert.NoError(t, err)

			if tt.expectResponse {
				assert.NotNil(t, resp)
			}
		})
	}
}

func TestProtoToModel(t *testing.T) {
	tests := []struct {
		name        string
		input       *pb.Metric
		expectError bool
		expectModel models.Metrics
		errorMsg    string
	}{
		{
			name:        "nil metric",
			input:       nil,
			expectError: true,
			errorMsg:    "nil metric",
		},
		{
			name: "empty id",
			input: &pb.Metric{
				Id:    "",
				Type:  pb.Metric_GAUGE,
				Value: 1.0,
			},
			expectError: true,
			errorMsg:    "id is empty",
		},
		{
			name: "valid gauge metric",
			input: &pb.Metric{
				Id:    "g1",
				Type:  pb.Metric_GAUGE,
				Value: 3.14,
			},
			expectError: false,
			expectModel: models.Metrics{
				ID:    "g1",
				MType: models.Gauge,
				Value: floatPtr(3.14),
			},
		},
		{
			name: "valid counter metric",
			input: &pb.Metric{
				Id:    "c1",
				Type:  pb.Metric_COUNTER,
				Delta: 42,
			},
			expectError: false,
			expectModel: models.Metrics{
				ID:    "c1",
				MType: models.Counter,
				Delta: intPtr(42),
			},
		},
		{
			name: "unknown metric type",
			input: &pb.Metric{
				Id:   "x1",
				Type: pb.Metric_UNKNOWN,
			},
			expectError: true,
			errorMsg:    "unknown metric type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			result, err := protoToModel(tt.input)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expectModel.ID, result.ID)
			assert.Equal(t, tt.expectModel.MType, result.MType)

			if tt.expectModel.Value != nil {
				assert.NotNil(t, result.Value)
				assert.Equal(t, *tt.expectModel.Value, *result.Value)
			}

			if tt.expectModel.Delta != nil {
				assert.NotNil(t, result.Delta)
				assert.Equal(t, *tt.expectModel.Delta, *result.Delta)
			}
		})
	}
}

func TestApiErrToStatus(t *testing.T) {
	tests := []struct {
		name       string
		input      *api.APIError
		expectCode codes.Code
		expectMsg  string
	}{
		{
			name:       "nil error converts to internal",
			input:      nil,
			expectCode: codes.Internal,
			expectMsg:  "internal server error",
		},
		{
			name: "bad request",
			input: &api.APIError{
				Code:    http.StatusBadRequest,
				Message: "bad request msg",
			},
			expectCode: codes.InvalidArgument,
			expectMsg:  "bad request msg",
		},
		{
			name: "unprocessable entity",
			input: &api.APIError{
				Code:    http.StatusUnprocessableEntity,
				Message: "unprocessable msg",
			},
			expectCode: codes.InvalidArgument,
			expectMsg:  "unprocessable msg",
		},
		{
			name: "not found",
			input: &api.APIError{
				Code:    http.StatusNotFound,
				Message: "not found msg",
			},
			expectCode: codes.NotFound,
			expectMsg:  "not found msg",
		},
		{
			name: "forbidden",
			input: &api.APIError{
				Code:    http.StatusForbidden,
				Message: "forbidden msg",
			},
			expectCode: codes.PermissionDenied,
			expectMsg:  "forbidden msg",
		},
		{
			name: "unauthorized",
			input: &api.APIError{
				Code:    http.StatusUnauthorized,
				Message: "unauth msg",
			},
			expectCode: codes.Unauthenticated,
			expectMsg:  "unauth msg",
		},
		{
			name: "method not allowed",
			input: &api.APIError{
				Code:    http.StatusMethodNotAllowed,
				Message: "method msg",
			},
			expectCode: codes.Unimplemented,
			expectMsg:  "method msg",
		},
		{
			name: "internal server error",
			input: &api.APIError{
				Code:    http.StatusInternalServerError,
				Message: "internal msg",
			},
			expectCode: codes.Internal,
			expectMsg:  "internal msg",
		},
		{
			name: "generic server error >= 500",
			input: &api.APIError{
				Code:    502,
				Message: "gateway error",
			},
			expectCode: codes.Internal,
			expectMsg:  "gateway error",
		},
		{
			name: "unknown client error",
			input: &api.APIError{
				Code:    418,
				Message: "teapot",
			},
			expectCode: codes.Unknown,
			expectMsg:  "teapot",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := apiErrToStatus(tt.input)

			st, ok := status.FromError(err)
			assert.True(t, ok)
			assert.Equal(t, tt.expectCode, st.Code())
			assert.Contains(t, st.Message(), tt.expectMsg)
		})
	}
}
