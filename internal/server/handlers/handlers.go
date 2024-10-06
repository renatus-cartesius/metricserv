package handlers

import (
	"net/http"
	"slices"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/renatus-cartesius/metricserv/internal/metrics"
	"github.com/renatus-cartesius/metricserv/internal/storage"
)

type ServerHandler struct {
	storage storage.Storage
}

func NewServerHandler(storage storage.Storage) *ServerHandler {
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

	w.Write([]byte(value))
}

func (srv ServerHandler) AllMetrics(w http.ResponseWriter, r *http.Request) {
	allMetrics, err := srv.storage.ListAll()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	body := ""
	for _, v := range allMetrics {
		body += v.String() + "\n"
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(body))
}
