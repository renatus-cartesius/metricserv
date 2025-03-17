package pb

import (
	"context"
	"fmt"
	"github.com/renatus-cartesius/metricserv/pkg/encryption"
	"github.com/renatus-cartesius/metricserv/pkg/logger"
	"github.com/renatus-cartesius/metricserv/pkg/metrics"
	"github.com/renatus-cartesius/metricserv/pkg/proto"
	"github.com/renatus-cartesius/metricserv/pkg/storage"
	"go.uber.org/zap"
	"net"
	"strconv"
)

type Server struct {
	proto.UnimplementedMetricsServiceServer

	TrustedSubnet *net.IPNet
	Storage       storage.Storager
	EncProcessor  encryption.Processor
}

func (s *Server) AddMetric(ctx context.Context, in *proto.AddMetricRequest) (*proto.AddMetricResponse, error) {

	var metric metrics.Metric
	var response proto.AddMetricResponse
	var metricType string

	switch in.Metric.Type {
	case proto.MetricType_COUNTER:

		metricType = metrics.TypeCounter
		value, err := strconv.ParseInt(in.Metric.Value, 10, 64)
		if err != nil {
			response.Error = fmt.Sprintf("error when parsing int64: %v", in.Metric.Value)
			return &response, err
		}

		metric = metrics.NewCounter(in.MetricID, value)
	case proto.MetricType_GAUGE:
		metricType = metrics.TypeGauge
		value, err := strconv.ParseFloat(in.Metric.Value, 32)

		if err != nil {
			response.Error = fmt.Sprintf("error when parsing float32: %v", in.Metric.Value)
			return &response, err
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

func (s *Server) GetMetric(ctx context.Context, in *proto.GetMetricRequest) (*proto.GetMetricResponse, error) {
	var response proto.GetMetricResponse
	var err error

	switch in.Type {
	case proto.MetricType_COUNTER:
		response.Value, err = s.Storage.GetValue(ctx, metrics.TypeCounter, in.MetricID)

		if err != nil {
			response.Error = fmt.Sprintf("error when getting counter metric: %v", in.MetricID)
			return &response, err
		}

	case proto.MetricType_GAUGE:
		response.Value, err = s.Storage.GetValue(ctx, metrics.TypeGauge, in.MetricID)

		if err != nil {
			response.Error = fmt.Sprintf("error when getting gauge metric: %v", in.MetricID)
			return &response, err
		}
	}

	return &response, err

}
