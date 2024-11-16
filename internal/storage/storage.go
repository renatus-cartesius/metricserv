package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
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
	Add(context.Context, string, metrics.Metric) error
	ListAll(context.Context) (map[string]metrics.Metric, error)
	CheckMetric(context.Context, string) bool
	Update(context.Context, string, string, any) error
	GetValue(context.Context, string, string) string
	Save(context.Context) error
	Load(context.Context) error
	Ping(context.Context) bool
	Close(context.Context)
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

func (s *MemStorage) Update(ctx context.Context, mtype, id string, value any) error {
	s.mx.Lock()
	defer s.mx.Unlock()

	metric := s.Metrics[id]
	if metric.GetType() != mtype {
		return ErrWrongUpdateType
	}
	metric.Change(value)

	return nil
}

func (s *MemStorage) Add(ctx context.Context, id string, metric metrics.Metric) error {
	s.mx.Lock()
	defer s.mx.Unlock()
	s.Metrics[id] = metric
	return nil
}

func (s *MemStorage) CheckMetric(ctx context.Context, id string) bool {
	// TODO: need to add check of metric type
	s.mx.RLock()
	defer s.mx.RUnlock()
	_, ok := s.Metrics[id]
	return ok
}

func (s *MemStorage) ListAll(ctx context.Context) (map[string]metrics.Metric, error) {
	s.mx.RLock()
	defer s.mx.RUnlock()
	return s.Metrics, nil
}

func (s *MemStorage) GetValue(ctx context.Context, mtype, id string) string {
	s.mx.RLock()
	defer s.mx.RUnlock()
	metric := s.Metrics[id]
	if metric.GetType() != mtype {
		return ""
	}
	return metric.GetValue()
}

func (s *MemStorage) Load(ctx context.Context) error {

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

		switch m["type"].(string) {
		case metrics.TypeCounter:
			// fmt.Println(m)
			counter := metrics.NewCounter(m["id"].(string), int64(m["value"].(float64)))
			s.Add(ctx, m["id"].(string), counter)
		case metrics.TypeGauge:
			gauge := metrics.NewGauge(m["id"].(string), m["value"].(float64))
			// fmt.Println(m)
			s.Add(ctx, m["id"].(string), gauge)
		}

	}

	logger.Log.Info(
		"succesfully loaded storage from file",
		zap.Int("metricsCount", len(s.Metrics)),
	)

	return nil
}

func (s *MemStorage) Save(ctx context.Context) error {

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

func (s *MemStorage) Ping(ctx context.Context) bool {
	return true
}

func (s *MemStorage) Close(ctx context.Context) {
}

type PGStorage struct {
	Storager
	db *sql.DB
}

func NewPGStorage(db *sql.DB) (Storager, error) {
	return &PGStorage{
		db: db,
	}, nil
}

func (pgs *PGStorage) Close(ctx context.Context) {
	if err := pgs.db.Close(); err != nil {
		log.Fatalln(err)
	}
}

func (pgs *PGStorage) Ping(ctx context.Context) bool {
	if err := pgs.db.PingContext(ctx); err != nil {
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

func (pgs *PGStorage) Add(ctx context.Context, id string, metric metrics.Metric) error {
	_, err := pgs.db.ExecContext(ctx, "INSERT INTO metrics (id, type, value) VALUES ($1, $2, $3)", id, metric.GetType(), metric.GetValue())
	if err != nil {
		logger.Log.Error(
			"error on inserting metric to db",
		)
		return err
	}
	return nil
}
func (pgs *PGStorage) ListAll(ctx context.Context) (map[string]metrics.Metric, error) {
	return nil, nil
}
func (pgs *PGStorage) CheckMetric(ctx context.Context, id string) bool {
	row := pgs.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM metrics WHERE id = $1", id)
	var count int
	row.Scan(&count)
	if count != 0 {
		return true
	}
	if err := row.Err(); err != nil {
		logger.Log.Error(
			"error on checking metric in db",
			zap.Error(err),
		)
	}
	return false
}
func (pgs *PGStorage) Update(ctx context.Context, mtype, id string, value any) error {

	// TODO: remove this workaround
	if mtype == metrics.TypeCounter {
		currentValue, err := strconv.ParseInt(pgs.GetValue(ctx, mtype, id), 10, 64)
		if err != nil {
			return err
		}

		value = value.(int64) + currentValue
	}

	_, err := pgs.db.ExecContext(ctx, "UPDATE metrics SET value = $1 where id = $2 AND type = $3", value, id, mtype)
	if err != nil {
		logger.Log.Error(
			"error on inserting metric to db",
		)
		return err
	}
	return nil
}
func (pgs *PGStorage) GetValue(ctx context.Context, mtype, id string) (res string) {
	row := pgs.db.QueryRowContext(ctx, "SELECT value FROM metrics WHERE id = $1 and type = $2", id, mtype)

	var value float64

	row.Scan(&value)
	if err := row.Err(); err != nil {
		logger.Log.Error(
			"error on getting metric from db",
			zap.Error(err),
		)
	}

	switch mtype {
	case metrics.TypeCounter:
		res = fmt.Sprintf("%v", int64(value))
	case metrics.TypeGauge:
		res = fmt.Sprintf("%v", value)
	}

	return
}
func (pgs *PGStorage) Save(ctx context.Context) error {
	return nil
}
func (pgs *PGStorage) Load(ctx context.Context) error {
	return nil
}
