// Package metrics providing common Metric interface and some its implementations (GaugeMetric, CounterMetric)
package metrics

const (
	TypeGauge   = "gauge"
	TypeCounter = "counter"
)

var (
	AllowedTypes = []string{
		TypeCounter,
		TypeGauge,
	}
)

type Metric interface {
	String() string
	Change(interface{}) error
	GetValue() string
	GetType() string
	GetID() string
}
