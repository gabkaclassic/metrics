package grpcserver

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"log/slog"

	models "github.com/gabkaclassic/metrics/internal/model"
	pb "github.com/gabkaclassic/metrics/internal/proto"
	"github.com/gabkaclassic/metrics/internal/service"
	api "github.com/gabkaclassic/metrics/pkg/error"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MetricsGRPCServer implements pb.MetricsServer using existing MetricsService.
type (
	MetricsGRPCServer struct {
		pb.UnimplementedMetricsServer
		svc service.MetricsService
	}

	GRPCServer struct {
		srv *grpc.Server
	}
)

// NewMetricsGRPCServer creates a new MetricsGRPCServer instance.
//
// svc: Business service responsible for metric processing.
//
// Returns:
//   - *MetricsGRPCServer: Ready-to-register gRPC server implementation.
func NewMetricsGRPCServer(svc service.MetricsService) *MetricsGRPCServer {
	return &MetricsGRPCServer{svc: svc}
}

// UpdateMetrics handles incoming metric batches and persists them via service layer.
//
// ctx: Incoming gRPC context containing metadata.
// req: Batch of metrics to update.
//
// Returns:
//   - *pb.UpdateMetricsResponse: Empty response on success.
//   - error: gRPC status error on validation or persistence failure.
func (s *MetricsGRPCServer) UpdateMetrics(ctx context.Context, req *pb.UpdateMetricsRequest) (*pb.UpdateMetricsResponse, error) {
	if req == nil || len(req.Metrics) == 0 {
		return &pb.UpdateMetricsResponse{}, nil
	}

	metrics := make([]models.Metrics, 0, len(req.Metrics))

	for i, pm := range req.Metrics {
		m, err := protoToModel(pm)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("metric[%d]: %v", i, err))
		}
		metrics = append(metrics, m)
	}

	if err := s.svc.SaveAll(ctx, metrics); err != nil {
		return nil, apiErrToStatus(err)
	}

	slog.Debug("UpdateMetrics completed", slog.Int("count", len(metrics)))
	return &pb.UpdateMetricsResponse{}, nil
}

// protoToModel converts pb.Metric to models.Metrics with validation.
//
// pm: Incoming protobuf metric.
//
// Returns:
//   - models.Metrics: Converted metric.
//   - error: Validation error for invalid fields.
func protoToModel(pm *pb.Metric) (models.Metrics, error) {
	if pm == nil {
		return models.Metrics{}, fmt.Errorf("nil metric")
	}

	if pm.Id == "" {
		return models.Metrics{}, fmt.Errorf("id is empty")
	}

	result := models.Metrics{
		ID: pm.Id,
	}

	switch pm.Type {
	case pb.Metric_GAUGE:
		value := pm.Value
		result.MType = models.Gauge
		result.Value = &value
	case pb.Metric_COUNTER:
		delta := pm.Delta
		result.MType = models.Counter
		result.Delta = &delta
	default:
		return models.Metrics{}, fmt.Errorf("unknown metric type: %v", pm.Type)
	}

	return result, nil
}

// apiErrToStatus converts APIError to gRPC status error.
//
// ae: Service-layer API error.
//
// Returns:
//   - error: Corresponding gRPC status error.
//
// apiErrToStatus converts APIError to gRPC status error.
//
// ae: Service-layer API error.
//
// Returns:
//   - error: Corresponding gRPC status error.
func apiErrToStatus(ae *api.APIError) error {
	if ae == nil {
		return status.Error(codes.Internal, "internal server error")
	}

	switch ae.Code {
	case http.StatusBadRequest:
		return status.Error(codes.InvalidArgument, ae.Message)
	case http.StatusUnprocessableEntity:
		return status.Error(codes.InvalidArgument, ae.Message)
	case http.StatusNotFound:
		return status.Error(codes.NotFound, ae.Message)
	case http.StatusForbidden:
		return status.Error(codes.PermissionDenied, ae.Message)
	case http.StatusUnauthorized:
		return status.Error(codes.Unauthenticated, ae.Message)
	case http.StatusMethodNotAllowed:
		return status.Error(codes.Unimplemented, ae.Message)
	case http.StatusInternalServerError:
		return status.Error(codes.Internal, ae.Message)
	default:
		if ae.Code >= 500 {
			return status.Error(codes.Internal, ae.Message)
		}
		return status.Error(codes.Unknown, ae.Message)
	}
}

func SetupGRPCServer(
	metricsService service.MetricsService,
	cidr string,
) (*GRPCServer, error) {

	trustInterceptor, err := trustAddressUnaryInterceptor(cidr)
	if err != nil {
		return nil, err
	}

	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			auditUnaryInterceptor(),
			trustInterceptor,
			loggingUnaryInterceptor(),
		),
	)

	pb.RegisterMetricsServer(server, NewMetricsGRPCServer(metricsService))

	return &GRPCServer{srv: server}, nil
}

func (g *GRPCServer) Run(ctx context.Context, addr string) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		slog.Error("failed to listen gRPC", "addr", addr, "error", err)
		return
	}

	serveDone := make(chan struct{})

	go func() {
		defer close(serveDone)

		if err := g.srv.Serve(lis); err != nil {
			slog.Error("gRPC server serve error", "error", err)
		}
	}()

	slog.Info("gRPC server started", "addr", addr)

	<-ctx.Done()
	slog.Info("gRPC shutdown requested")

	g.srv.GracefulStop()
	<-serveDone

	slog.Info("gRPC server stopped gracefully")
}
