package storage

import (
	"context"
	"fmt"
	"testing"

	"github.com/renatus-cartesius/metricserv/pkg/metrics"
)

func TestMemStorageAdd(t *testing.T) {
	// Counter metric test
	counterID := "poll_count"
	counterValue := int64(12345678)

	counter := &metrics.CounterMetric{
		ID:    counterID,
		Value: counterValue,
	}

	// Gauge metric test
	gaugeID := "cpu_usage"
	gaugeValue := float64(12345.678)

	gauge := &metrics.GaugeMetric{
		ID:    gaugeID,
		Value: gaugeValue,
	}

	storage, err := NewMemStorage("./storage")
	if err != nil {
		t.Fatalf("error on creating new storage\n")
	}

	storage.Add(context.Background(), counterID, counter)
	storage.Add(context.Background(), gaugeID, gauge)

	metrics, err := storage.ListAll(context.Background())
	if err != nil {
		t.Fatalf("error on listings storage %s", err)
	}

	fmt.Println("DEBUG:", metrics)
}
