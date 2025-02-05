package metrics

import "fmt"

type GaugeMetric struct {
	ID    string  `json:"id"`
	Value float64 `json:"value"`
	Type  string  `json:"type"`
}

func NewGauge(id string, value float64) *GaugeMetric {
	return &GaugeMetric{
		ID:    id,
		Value: value,
		Type:  TypeGauge,
	}
}

func (m *GaugeMetric) GetID() string {
	return m.ID
}

func (m *GaugeMetric) Change(value interface{}) error {
	m.Value = value.(float64)
	return nil
}
func (m GaugeMetric) String() string {
	return fmt.Sprintf("%s:%s:%f", TypeGauge, m.ID, m.Value)
}

func (m *GaugeMetric) GetValue() string {
	return fmt.Sprintf("%v", m.Value)
}

func (m *GaugeMetric) GetType() string {
	return TypeGauge
}
