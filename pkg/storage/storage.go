// Package storage implements Storager interface for using it as a metrics.Metric storage backend.
package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/renatus-cartesius/metricserv/pkg/logger"
	"github.com/renatus-cartesius/metricserv/pkg/metrics"
	"github.com/renatus-cartesius/metricserv/pkg/utils"
	"go.uber.org/zap"
)

var (
	ErrWrongUpdateType = errors.New("wrong type when updating metric in storage")
	ErrWrongGetType    = errors.New("wrong type when getting metric from storage")
	ErrEmptyMemStorage = errors.New("memstorage is not initialized")
)

// Storager represents data repository and working with some kind of underlying datastore (memory, file, dbms) and exposes CRUD operations. Data can be Load from underlying datastore on init phase.
type Storager interface {
	// Add Adds new metric to storage.
	Add(context.Context, string, metrics.Metric) error

	// ListAll Listing all metrics in storage.
	ListAll(context.Context) (map[string]metrics.Metric, error)

	// CheckMetric checking if metric is in storage by it`s id.
	CheckMetric(context.Context, string) (bool, error)

	// Update updates already added to storage metric.
	Update(context.Context, string, string, any) error

	// GetValue returning value of metric as strings.
	GetValue(context.Context, string, string) (string, error)

	// Save saving all metrics to underlying datastore
	Save(context.Context) error

	// Load loads all metrics from underlying datastore.
	Load(context.Context) error

	// Ping checks if underlying datastore is available.
	Ping(context.Context) error

	// Close free underlying datastore.
	Close() error
}

// MemStorage implements Storager using memory and files and underlying datastore.
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

func (s *MemStorage) CheckMetric(ctx context.Context, id string) (bool, error) {
	// TODO: need to add check of metric type
	s.mx.RLock()
	defer s.mx.RUnlock()
	_, ok := s.Metrics[id]
	return ok, nil
}

func (s *MemStorage) ListAll(ctx context.Context) (map[string]metrics.Metric, error) {
	s.mx.RLock()
	defer s.mx.RUnlock()
	return s.Metrics, nil
}

func (s *MemStorage) GetValue(ctx context.Context, mtype, id string) (string, error) {
	s.mx.RLock()
	defer s.mx.RUnlock()
	metric := s.Metrics[id]
	if metric.GetType() != mtype {
		return "", nil
	}
	return metric.GetValue(), nil
}

// Loading metrics from file to MemStorage.Metrics
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

func (s *MemStorage) Ping(ctx context.Context) error {
	if s.Metrics == nil {
		return ErrEmptyMemStorage
	}
	return nil
}

func (s *MemStorage) Close() error {
	return nil
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

func (pgs *PGStorage) Close() error {
	return pgs.db.Close()
}

func (pgs *PGStorage) Ping(ctx context.Context) error {
	return pgs.db.PingContext(ctx)
}

func (pgs *PGStorage) Add(ctx context.Context, id string, metric metrics.Metric) error {
	_, err := pgs.db.ExecContext(ctx, "INSERT INTO metrics (id, type, value) VALUES ($1, $2, $3)", id, metric.GetType(), metric.GetValue())
	return err
}

func (pgs *PGStorage) ListAll(ctx context.Context) (map[string]metrics.Metric, error) {
	return nil, nil
}

func (pgs *PGStorage) CheckMetric(ctx context.Context, id string) (bool, error) {
	row := pgs.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM metrics WHERE id = $1", id)
	var count int
	row.Scan(&count)

	if count != 0 {
		return true, nil
	}

	err := row.Err()
	return false, err
}
func (pgs *PGStorage) Update(ctx context.Context, mtype, id string, value any) error {

	// TODO: remove this workaround
	if mtype == metrics.TypeCounter {

		stringValue, err := pgs.GetValue(ctx, mtype, id)
		if err != nil {
			return err
		}

		currentValue, err := strconv.ParseInt(stringValue, 10, 64)
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
func (pgs *PGStorage) GetValue(ctx context.Context, mtype, id string) (string, error) {
	row := pgs.db.QueryRowContext(ctx, "SELECT value FROM metrics WHERE id = $1 and type = $2", id, mtype)

	var value float64

	row.Scan(&value)
	if err := row.Err(); err != nil {
		return "", err
	}

	switch mtype {
	case metrics.TypeCounter:
		return fmt.Sprintf("%v", int64(value)), nil
	case metrics.TypeGauge:
		return fmt.Sprintf("%v", value), nil
	default:
		return "", ErrWrongGetType
	}
}
func (pgs *PGStorage) Save(ctx context.Context) error {
	return nil
}
func (pgs *PGStorage) Load(ctx context.Context) error {
	return nil
}
