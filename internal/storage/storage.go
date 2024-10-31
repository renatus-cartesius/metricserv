package storage

import (
	"encoding/json"
	"errors"
	"os"
	"sync"

	"github.com/renatus-cartesius/metricserv/internal/logger"
	"github.com/renatus-cartesius/metricserv/internal/metrics"
	"github.com/renatus-cartesius/metricserv/internal/utils"
	"go.uber.org/zap"
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
	Save() error
	Load() error
}

type MemStorage struct {
	mx       sync.RWMutex
	Metrics  map[string]metrics.Metric
	savePath string
}

func NewMemStorage(savePath string) Storager {
	return &MemStorage{
		Metrics:  make(map[string]metrics.Metric, 0),
		savePath: savePath,
	}
}

func (s *MemStorage) Update(mtype, name string, value any) error {
	s.mx.Lock()
	defer s.mx.Unlock()

	metric := s.Metrics[name]
	if metric.GetType() != mtype {
		return ErrWrongUpdateType
	}
	metric.Change(value)

	return nil
}

func (s *MemStorage) Add(name string, metric metrics.Metric) error {
	s.mx.Lock()
	defer s.mx.Unlock()
	s.Metrics[name] = metric
	return nil
}

func (s *MemStorage) CheckMetric(name string) bool {
	// TODO: need to add check of metric type
	s.mx.RLock()
	defer s.mx.RUnlock()
	_, ok := s.Metrics[name]
	return ok
}

func (s *MemStorage) ListAll() (map[string]metrics.Metric, error) {
	s.mx.RLock()
	defer s.mx.RUnlock()
	metrics := s.Metrics
	return metrics, nil
}

func (s *MemStorage) GetValue(mtype, name string) string {
	s.mx.RLock()
	defer s.mx.RUnlock()
	metric := s.Metrics[name]
	if metric.GetType() != mtype {
		return ""
	}
	return metric.GetValue()
}

func (s *MemStorage) Load() error {

	fileInfo, err := os.Stat(s.savePath)
	if err != nil {

		if os.IsNotExist(err) {
			return nil
		}

		if fileInfo.Size() == 0 {
			return nil
		}

		logger.Log.Error(
			"error on getting info about file for loading storage",
			zap.Error(err),
		)
		return err
	}

	file, err := os.OpenFile(s.savePath, os.O_RDONLY, 0666)
	if err != nil {
		logger.Log.Error(
			"error on opening file for loading storage",
			zap.Error(err),
		)
		return err
	}
	defer utils.CloseFile(file)

	s.mx.Lock()
	defer s.mx.Unlock()

	if err := json.NewDecoder(file).Decode(s); err != nil {
		logger.Log.Error(
			"error on unmarshaling file to object for loading storage",
			zap.Error(err),
		)
		return err
	}

	logger.Log.Debug(
		"succesfully loaded storage from file",
		zap.Int("metricsCount", len(s.Metrics)),
	)

	return nil
}

func (s *MemStorage) Save() error {

	logger.Log.Info(
		"saving storage to file",
	)

	file, err := os.OpenFile(s.savePath, os.O_WRONLY|os.O_CREATE, 0666)
	defer utils.CloseFile(file)

	if err != nil {
		logger.Log.Error(
			"error on opening or creating storage file",
		)
		return err
	}

	if err := json.NewEncoder(file).Encode(s); err != nil {
		logger.Log.Error(
			"error on marshalling storage for saving",
		)
		return err
	}

	return nil
}
