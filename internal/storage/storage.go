package storage

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
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
	Ping() bool
}

type MemStorage struct {
	Storager
	mx       sync.RWMutex
	Metrics  map[string]metrics.Metric `json:"metrics"`
	savePath string
}

func NewMemStorage(savePath string) (Storager, error) {
	return &MemStorage{
		Metrics:  make(map[string]metrics.Metric, 0),
		savePath: savePath,
	}, nil
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
	return s.Metrics, nil
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

	logger.Log.Info(
		"loading storage from file",
		zap.String("filepath", s.savePath),
	)

	file, err := os.OpenFile(s.savePath, os.O_RDONLY, 0666)
	if err != nil {
		logger.Log.Error(
			"error on opening file for loading storage",
			zap.Error(err),
		)
		return err
	}
	defer utils.SafeClose(file)

	// s.mx.Lock()
	// defer s.mx.Unlock()

	// var tmp map[string]models.AbstractMetric
	var tmp interface{}

	if err := json.NewDecoder(file).Decode(&tmp); err != nil {
		logger.Log.Error(
			"error on unmarshaling file to object for loading storage",
			zap.Error(err),
		)
		return err
	}

	for _, v := range tmp.(map[string]interface{})["metrics"].(map[string]interface{}) {
		m := v.(map[string]interface{})
		// m := v.(models.AbstractMetric)

		switch m["type"].(string) {
		case metrics.TypeCounter:
			// fmt.Println(m)
			counter := metrics.NewCounter(m["name"].(string), int64(m["value"].(float64)))
			s.Add(m["name"].(string), counter)
		case metrics.TypeGauge:
			gauge := metrics.NewGauge(m["name"].(string), m["value"].(float64))
			// fmt.Println(m)
			s.Add(m["name"].(string), gauge)
		}

	}

	logger.Log.Info(
		"succesfully loaded storage from file",
		zap.Int("metricsCount", len(s.Metrics)),
	)

	return nil
}

func (s *MemStorage) Save() error {

	logger.Log.Info(
		"saving storage to file",
	)

	file, err := os.OpenFile(s.savePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	defer utils.SafeClose(file)

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

func (s *MemStorage) Ping() bool {
	return true
}

type PGStorage struct {
	Storager
	db *sql.DB
}

func NewPGStorage(dsn string) (Storager, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalln(err)
	}

	return &PGStorage{
		db: db,
	}, nil
}

func (pgs *PGStorage) Close() {
	if err := pgs.db.Close(); err != nil {
		log.Fatalln(err)
	}
}

func (pgs *PGStorage) Ping() bool {
	if err := pgs.db.Ping(); err != nil {
		logger.Log.Error(
			"error on ping postgresql database server",
			zap.Error(err),
		)
		return false
	}
	logger.Log.Debug(
		"connection to postgresql database server is alive",
	)
	return true
}

func (pgs *PGStorage) Add(string, metrics.Metric) error {
	return nil
}
func (pgs *PGStorage) ListAll() (map[string]metrics.Metric, error) {
	return nil, nil
}
func (pgs *PGStorage) CheckMetric(string) bool {
	return true
}
func (pgs *PGStorage) Update(string, string, any) error {
	return nil
}
func (pgs *PGStorage) GetValue(string, string) string {
	return ""
}
func (pgs *PGStorage) Save() error {
	return nil
}
func (pgs *PGStorage) Load() error {
	return nil
}
