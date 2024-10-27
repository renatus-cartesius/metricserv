package storage

import (
	"errors"
	"sync"

	"github.com/renatus-cartesius/metricserv/internal/metrics"
)

var (
	ErrWrongUpdateType = errors.New("wrong type when updating")
)

type Storager interface {
	Add(string, metrics.Metric) error
	ListAll() (map[string]metrics.Metric, error)
	CheckMetric(string) bool
	Update(string, string, any) error
	GetValue(string, string) string
}

type MemStorage struct {
	mx      sync.RWMutex
	metrics map[string]metrics.Metric
}

func NewMemStorage() Storager {
	return &MemStorage{
		metrics: make(map[string]metrics.Metric, 0),
	}
}

func (s *MemStorage) Update(mtype, name string, value any) error {
	s.mx.Lock()
	defer s.mx.Unlock()

	metric := s.metrics[name]
	if metric.GetType() != mtype {
		return ErrWrongUpdateType
	}
	metric.Change(value)

	return nil
}

func (s *MemStorage) Add(name string, metric metrics.Metric) error {
	s.mx.Lock()
	defer s.mx.Unlock()
	s.metrics[name] = metric
	return nil
}

func (s *MemStorage) CheckMetric(name string) bool {
	// TODO: need to add check of metric type
	s.mx.RLock()
	defer s.mx.RUnlock()
	_, ok := s.metrics[name]
	return ok
}

func (s *MemStorage) ListAll() (map[string]metrics.Metric, error) {
	s.mx.RLock()
	defer s.mx.RUnlock()
	metrics := s.metrics
	return metrics, nil
}

func (s *MemStorage) GetValue(mtype, name string) string {
	s.mx.RLock()
	defer s.mx.RUnlock()
	metric := s.metrics[name]
	if metric.GetType() != mtype {
		return ""
	}
	return metric.GetValue()
}
