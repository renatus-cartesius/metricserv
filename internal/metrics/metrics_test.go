package metrics

import (
	"fmt"
	"testing"
)

func TestNewGaugeMetric(t *testing.T) {
	metricName := "cpu_usage"
	metricValue := float64(10)

	metric := &GaugeMetric{
		Name:  metricName,
		Value: metricValue,
	}

	if fmt.Sprint(metric) != "gauge:cpu_usage:10.000000" {
		t.Errorf("Error on converting metric to string")
	}

	fmt.Println(metric)
}

func TestNewCounterMetric(t *testing.T) {
	metricName := "poll_count"
	metricValue := int64(12345678)

	metric := &CounterMetric{
		Name:  metricName,
		Value: metricValue,
	}

	if fmt.Sprint(metric) != "counter:poll_count:12345678" {
		t.Errorf("Error on converting metric to string")
	}

	fmt.Println(metric)
}
