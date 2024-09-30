package storage

import (
	"sync"

	"github.com/renatus-cartesius/metricserv/internal/metrics"
)

type MemStorage struct {
	mx      sync.RWMutex
	metrics map[string]metrics.Metric
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		metrics: make(map[string]metrics.Metric, 0),
	}
}

func (s *MemStorage) Update(name string, value interface{}) error {
	s.mx.Lock()
	defer s.mx.Unlock()

	metric, _ := s.metrics[name]
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
