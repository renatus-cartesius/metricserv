package pb

import (
	"context"
	"github.com/renatus-cartesius/metricserv/pkg/api"
	"github.com/renatus-cartesius/metricserv/pkg/encryption"
	"github.com/renatus-cartesius/metricserv/pkg/logger"
	"github.com/renatus-cartesius/metricserv/pkg/metrics"
	"github.com/renatus-cartesius/metricserv/pkg/storage"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net"
	"strconv"
)

type Server struct {
	api.UnimplementedMetricsServiceServer

	TrustedSubnet *net.IPNet
	Storage       storage.Storager
	EncProcessor  encryption.Processor
}

func (s *Server) AddMetric(ctx context.Context, in *api.AddMetricRequest) (*api.AddMetricResponse, error) {

	var metric metrics.Metric
	var response api.AddMetricResponse
	var metricType string

	switch in.Metric.Type {
	case api.MetricType_COUNTER:

		metricType = metrics.TypeCounter
		value, err := strconv.ParseInt(in.Metric.Value, 10, 64)
		if err != nil {
			return &response, status.Errorf(codes.Internal, "error when parsing int64: %v", in.Metric.Value)
		}

		metric = metrics.NewCounter(in.MetricID, value)
	case api.MetricType_GAUGE:
		metricType = metrics.TypeGauge
		value, err := strconv.ParseFloat(in.Metric.Value, 32)

		if err != nil {
			return &response, status.Errorf(codes.Internal, "error when parsing float32: %v", in.Metric.Value)
		}

		metric = metrics.NewGauge(in.MetricID, value)
	}

	logger.Log.Info(
		"added metric",
		zap.String("metricID", in.MetricID),
		zap.String("type", metricType),
	)

	err := s.Storage.Add(ctx, metric.GetID(), metric)

	return &response, err
}

func (s *Server) GetMetric(ctx context.Context, in *api.GetMetricRequest) (*api.GetMetricResponse, error) {
	var response api.GetMetricResponse
	var err error

	switch in.Type {
	case api.MetricType_COUNTER:
		response.Value, err = s.Storage.GetValue(ctx, metrics.TypeCounter, in.MetricID)

		if err != nil {
			return &response, status.Errorf(codes.Internal, "error when getting counter metric: %v", in.MetricID)
		}

	case api.MetricType_GAUGE:
		response.Value, err = s.Storage.GetValue(ctx, metrics.TypeGauge, in.MetricID)

		if err != nil {
			return &response, status.Errorf(codes.Internal, "error when getting gauge metric: %v", in.MetricID)
		}
	}

	return &response, err

}
