package metrics

import "fmt"

type GaugeMetric struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
	Type  string  `json:"type"`
}

func NewGauge(name string, value float64) *GaugeMetric {
	return &GaugeMetric{
		Name:  name,
		Value: value,
		Type:  "gauge",
	}
}
func (m *GaugeMetric) Change(value interface{}) error {
	m.Value = value.(float64)
	return nil
}
func (m GaugeMetric) String() string {
	return fmt.Sprintf("gauge:%s:%f", m.Name, m.Value)
}

func (m *GaugeMetric) GetValue() string {
	return fmt.Sprintf("%v", m.Value)
}

func (m *GaugeMetric) GetType() string {
	return TypeGauge
}
