package storage

import (
	"github.com/renatus-cartesius/metricserv/internal/metrics"
)

type Storage interface {
	Add(*metrics.Metric) error
	ListAll() ([]*metrics.Metric, error)
}
