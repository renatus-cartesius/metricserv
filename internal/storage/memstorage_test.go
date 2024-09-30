package storage

import (
	"fmt"
	"testing"

	"github.com/renatus-cartesius/metricserv/internal/metrics"
)

func TestMemStorageAdd(t *testing.T) {
	// Counter metric test
	counterName := "poll_count"
	counterValue := int64(12345678)

	counter := &metrics.CounterMetric{
		Name:  counterName,
		Value: counterValue,
	}

	// Gauge metric test
	gaugeName := "cpu_usage"
	gaugeValue := float64(12345.678)

	gauge := &metrics.GaugeMetric{
		Name:  gaugeName,
		Value: gaugeValue,
	}

	storage := NewMemStorage()

	storage.Add(counterName, counter)
	storage.Add(gaugeName, gauge)

	metrics, _ := storage.ListAll()

	fmt.Println("DEBUG:", metrics)
}
