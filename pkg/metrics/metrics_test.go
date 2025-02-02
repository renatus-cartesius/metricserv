package metrics

import (
	"fmt"
	"testing"
)

func TestNewGaugeMetric(t *testing.T) {
	metricID := "cpu_usage"
	metricValue := float64(10)

	metric := &GaugeMetric{
		ID:    metricID,
		Value: metricValue,
	}

	if fmt.Sprint(metric) != "gauge:cpu_usage:10.000000" {
		t.Errorf("Error on converting metric to string")
	}

	fmt.Println(metric)
}

func TestNewCounterMetric(t *testing.T) {
	metricID := "poll_count"
	metricValue := int64(12345678)

	metric := &CounterMetric{
		ID:    metricID,
		Value: metricValue,
	}

	if fmt.Sprint(metric) != "counter:poll_count:12345678" {
		t.Errorf("Error on converting metric to string")
	}

	fmt.Println(metric)
}
