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
}
