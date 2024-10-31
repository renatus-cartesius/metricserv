package handlers

import (
	"encoding/json"
	"net/http"
	"slices"
	"strconv"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/renatus-cartesius/metricserv/internal/logger"
	"github.com/renatus-cartesius/metricserv/internal/metrics"
	"github.com/renatus-cartesius/metricserv/internal/server/models"
	"github.com/renatus-cartesius/metricserv/internal/storage"
)

type ServerHandler struct {
	storage storage.Storager
}

func NewServerHandler(storage storage.Storager) *ServerHandler {
	return &ServerHandler{
		storage: storage,
	}
}

func (srv ServerHandler) Update(w http.ResponseWriter, r *http.Request) {

	metricType := chi.URLParam(r, "type")
	metricName := chi.URLParam(r, "name")
	metricValue := chi.URLParam(r, "value")

	if !slices.Contains(metrics.AllowedTypes, metricType) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	switch metricType {
	case metrics.TypeCounter:
		value, err := strconv.ParseInt(metricValue, 10, 64)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if !srv.storage.CheckMetric(metricName) {
			metric := &metrics.CounterMetric{
				Name:  metricName,
				Value: int64(0),
			}
			err := srv.storage.Add(metricName, metric)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		err = srv.storage.Update(metricType, metricName, value)
		if err != nil {
			if err == storage.ErrWrongUpdateType {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

	case metrics.TypeGauge:
		value, err := strconv.ParseFloat(metricValue, 64)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if !srv.storage.CheckMetric(metricName) {
			metric := &metrics.GaugeMetric{
				Name:  metricName,
				Value: float64(0),
			}
			err := srv.storage.Add(metricName, metric)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		err = srv.storage.Update(metricType, metricName, value)
		if err != nil {
			if err == storage.ErrWrongUpdateType {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	default:
		w.WriteHeader(http.StatusNotImplemented)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
}

func (srv ServerHandler) GetValue(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "type")
	metricName := chi.URLParam(r, "name")

	if !srv.storage.CheckMetric(metricName) {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	value := srv.storage.GetValue(metricType, metricName)
	if value == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(value))
}

func (srv ServerHandler) GetValueJSON(w http.ResponseWriter, r *http.Request) {

	var err error

	var metric models.Metric

	if err = json.NewDecoder(r.Body).Decode(&metric); err != nil {
		logger.Log.Error(
			"error on unmarshaling request body",
			zap.Error(err),
		)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !srv.storage.CheckMetric(metric.ID) {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	value := srv.storage.GetValue(metric.MType, metric.ID)
	if value == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	switch metric.MType {
	case metrics.TypeCounter:
		metric.Delta = new(int64)
		*metric.Delta, _ = strconv.ParseInt(value, 10, 64)
	case metrics.TypeGauge:
		metric.Value = new(float64)
		*metric.Value, _ = strconv.ParseFloat(value, 64)
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	result, err := json.Marshal(metric)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(result)
}

func (srv ServerHandler) UpdateJSON(w http.ResponseWriter, r *http.Request) {

	var metric models.Metric

	if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
		logger.Log.Error(
			"error on unmarshaling request body",
			zap.Error(err),
		)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !slices.Contains(metrics.AllowedTypes, metric.MType) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	switch metric.MType {
	case metrics.TypeCounter:
		delta := metric.Delta

		if !srv.storage.CheckMetric(metric.ID) {
			newMetric := &metrics.CounterMetric{
				Name:  metric.ID,
				Value: int64(0),
			}
			err := srv.storage.Add(metric.ID, newMetric)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		err := srv.storage.Update(metric.MType, metric.ID, *delta)
		if err != nil {
			if err == storage.ErrWrongUpdateType {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		actualDelta, err := strconv.ParseInt(srv.storage.GetValue(metric.MType, metric.ID), 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		*metric.Delta = actualDelta

	case metrics.TypeGauge:
		value := metric.Value

		if !srv.storage.CheckMetric(metric.ID) {
			newMetric := &metrics.GaugeMetric{
				Name:  metric.ID,
				Value: float64(0),
			}
			err := srv.storage.Add(metric.ID, newMetric)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		err := srv.storage.Update(metric.MType, metric.ID, *value)
		if err != nil {
			if err == storage.ErrWrongUpdateType {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

	default:
		w.WriteHeader(http.StatusNotImplemented)
		return
	}

	result, err := json.Marshal(metric)
	if err != nil {
		logger.Log.Error(
			"error on marshaling result",
			zap.String("err", err.Error()),
		)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(result)
}

func (srv ServerHandler) AllMetrics(w http.ResponseWriter, r *http.Request) {
	allMetrics, err := srv.storage.ListAll()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	body := ""
	for _, v := range allMetrics {
		body += "<p>" + v.String() + "</p>"
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(body))
}
