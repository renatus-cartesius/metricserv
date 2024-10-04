package handlers

import (
	"net/http"
	"slices"
	"strconv"

	"github.com/renatus-cartesius/metricserv/internal/metrics"
	"github.com/renatus-cartesius/metricserv/internal/storage"
)

var memStorage *storage.MemStorage

func init() {
	memStorage = storage.NewMemStorage()
}

func UpdateHandler(w http.ResponseWriter, r *http.Request) {

	if !slices.Contains(metrics.AllowedTypes, r.PathValue("type")) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	metricType := r.PathValue("type")
	metricName := r.PathValue("name")
	metricValue := r.PathValue("value")

	switch metricType {
	case metrics.TypeCounter:
		value, err := strconv.ParseInt(metricValue, 10, 64)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if !memStorage.CheckMetric(metricName) {
			metric := &metrics.CounterMetric{
				Name:  metricName,
				Value: value,
			}
			err := memStorage.Add(metricName, metric)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		err = memStorage.Update(metricName, value)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

	case metrics.TypeGauge:
		value, err := strconv.ParseFloat(r.PathValue("value"), 64)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if !memStorage.CheckMetric(metricName) {
			metric := &metrics.GaugeMetric{
				Name:  metricName,
				Value: value,
			}
			err := memStorage.Add(metricName, metric)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		err = memStorage.Update(metricName, value)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	default:
		// w.WriteHeader(http.StatusBadRequest)
		return
	}

	allMetrics, err := memStorage.ListAll()
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
