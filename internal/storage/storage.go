package storage

import (
	"errors"

	"github.com/renatus-cartesius/metricserv/internal/metrics"
)

var (
	ErrWrongUpdateType = errors.New("wrong type when updating")
)

type Storage interface {
	Add(string, metrics.Metric) error
	ListAll() (map[string]metrics.Metric, error)
	CheckMetric(string) bool
	Update(string, string, interface{}) error
	GetValue(string, string) string
}
