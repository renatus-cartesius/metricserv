package pb

import (
	"context"
	api2 "github.com/renatus-cartesius/metricserv/api"
	"github.com/renatus-cartesius/metricserv/pkg/encryption"
	"github.com/renatus-cartesius/metricserv/pkg/logger"
	"github.com/renatus-cartesius/metricserv/pkg/metrics"
	"github.com/renatus-cartesius/metricserv/pkg/storage"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"net"
	"strconv"
)

type Server struct {
	api2.UnimplementedMetricsServiceServer

	TrustedSubnet *net.IPNet
	Storage       storage.Storager
	EncProcessor  encryption.Processor
}

func (s *Server) AddMetric(ctx context.Context, in *api2.AddMetricRequest) (*emptypb.Empty, error) {

	var metric metrics.Metric
	var metricType string

	switch in.Metric.Type {
	case api2.MetricType_COUNTER:

		metricType = metrics.TypeCounter
		value, err := strconv.ParseInt(in.Metric.Value, 10, 64)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "error when parsing int64: %v", in.Metric.Value)
		}

		metric = metrics.NewCounter(in.MetricID, value)
	case api2.MetricType_GAUGE:
		metricType = metrics.TypeGauge
		value, err := strconv.ParseFloat(in.Metric.Value, 32)

		if err != nil {
			return nil, status.Errorf(codes.Internal, "error when parsing float32: %v", in.Metric.Value)
		}

		metric = metrics.NewGauge(in.MetricID, value)
	}

	logger.Log.Info(
		"added metric",
		zap.String("metricID", in.MetricID),
		zap.String("type", metricType),
	)

	err := s.Storage.Add(ctx, metric.GetID(), metric)

	return &emptypb.Empty{}, err
}

func (s *Server) GetMetric(ctx context.Context, in *api2.GetMetricRequest) (*api2.GetMetricResponse, error) {
	var response api2.GetMetricResponse
	var err error

	switch in.Type {
	case api2.MetricType_COUNTER:
		response.Value, err = s.Storage.GetValue(ctx, metrics.TypeCounter, in.MetricID)

		if err != nil {
			return nil, status.Errorf(codes.Internal, "error when getting counter metric: %v", in.MetricID)
		}

	case api2.MetricType_GAUGE:
		response.Value, err = s.Storage.GetValue(ctx, metrics.TypeGauge, in.MetricID)

		if err != nil {
			return nil, status.Errorf(codes.Internal, "error when getting gauge metric: %v", in.MetricID)
		}
	}

	return &response, err

}
