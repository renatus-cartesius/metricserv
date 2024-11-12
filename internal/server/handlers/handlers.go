package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strconv"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/renatus-cartesius/metricserv/internal/logger"
	"github.com/renatus-cartesius/metricserv/internal/metrics"
	"github.com/renatus-cartesius/metricserv/internal/server/middlewares"
	"github.com/renatus-cartesius/metricserv/internal/server/models"
	"github.com/renatus-cartesius/metricserv/internal/storage"
)

func Setup(r *chi.Mux, srv *ServerHandler) {

	r.Route("/", func(r chi.Router) {
		r.Get("/", middlewares.Gzipper(logger.RequestLogger(srv.AllMetrics)))
		r.Get("/ping", middlewares.Gzipper(logger.RequestLogger(srv.Ping)))
		r.Route("/value", func(r chi.Router) {
			r.Post("/", middlewares.Gzipper(logger.RequestLogger(srv.GetValueJSON)))
			r.Get("/{type}/{id}", middlewares.Gzipper(logger.RequestLogger(srv.GetValue)))
		})
		r.Route("/update", func(r chi.Router) {
			r.Post("/", middlewares.Gzipper(logger.RequestLogger(srv.UpdateJSON)))
			r.Post("/{type}/{id}/{value}", middlewares.Gzipper(logger.RequestLogger(srv.Update)))
		})
	})
}

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
	metricID := chi.URLParam(r, "id")
	metricValue := chi.URLParam(r, "value")

	if !slices.Contains(metrics.AllowedTypes, metricType) {
		logger.Log.Warn(
			"passed type is not allowed",
			zap.String("type", metricType),
		)
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

		if !srv.storage.CheckMetric(r.Context(), metricID) {
			metric := &metrics.CounterMetric{
				ID:    metricID,
				Value: int64(0),
			}
			err := srv.storage.Add(r.Context(), metricID, metric)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		err = srv.storage.Update(r.Context(), metricType, metricID, value)
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

		if !srv.storage.CheckMetric(r.Context(), metricID) {
			metric := &metrics.GaugeMetric{
				ID:    metricID,
				Value: float64(0),
			}
			err := srv.storage.Add(r.Context(), metricID, metric)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		err = srv.storage.Update(r.Context(), metricType, metricID, value)
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
	metricID := chi.URLParam(r, "id")

	if !srv.storage.CheckMetric(r.Context(), metricID) {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	value := srv.storage.GetValue(r.Context(), metricType, metricID)
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
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !srv.storage.CheckMetric(r.Context(), metric.ID) {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	value := srv.storage.GetValue(r.Context(), metric.MType, metric.ID)
	if value == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	switch metric.MType {
	case metrics.TypeCounter:
		metric.Delta = new(int64)
		*metric.Delta, err = strconv.ParseInt(value, 10, 64)
		if err != nil {
			logger.Log.Error(
				"error on parsing value",
				zap.Error(err),
			)
			return
		}
	case metrics.TypeGauge:
		metric.Value = new(float64)
		*metric.Value, err = strconv.ParseFloat(value, 64)
		if err != nil {
			logger.Log.Error(
				"error on parsing value",
				zap.Error(err),
			)
			return
		}
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
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !slices.Contains(metrics.AllowedTypes, metric.MType) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	switch metric.MType {
	case metrics.TypeCounter:
		delta := metric.Delta

		if !srv.storage.CheckMetric(r.Context(), metric.ID) {
			fmt.Println("Creating new counter", metric.ID)
			newMetric := metrics.NewCounter(metric.ID, int64(0))
			err := srv.storage.Add(r.Context(), metric.ID, newMetric)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		err := srv.storage.Update(r.Context(), metric.MType, metric.ID, *delta)
		if err != nil {
			if err == storage.ErrWrongUpdateType {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		actualDelta, err := strconv.ParseInt(srv.storage.GetValue(r.Context(), metric.MType, metric.ID), 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		*metric.Delta = actualDelta

	case metrics.TypeGauge:
		value := metric.Value

		if !srv.storage.CheckMetric(r.Context(), metric.ID) {
			newMetric := metrics.NewGauge(metric.ID, float64(0))
			err := srv.storage.Add(r.Context(), metric.ID, newMetric)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		err := srv.storage.Update(r.Context(), metric.MType, metric.ID, *value)
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
	allMetrics, err := srv.storage.ListAll(r.Context())
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

func (srv ServerHandler) Ping(w http.ResponseWriter, r *http.Request) {
	if ok := srv.storage.Ping(r.Context()); !ok {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)
}
