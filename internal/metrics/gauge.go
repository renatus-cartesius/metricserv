package metrics

import "fmt"

type GaugeMetric struct {
	Name  string
	Value float64
}

func (m *GaugeMetric) Change(value interface{}) error {
	m.Value = value.(float64)
	return nil
}
func (m *GaugeMetric) String() string {
	return fmt.Sprintf("gauge:%s:%f", m.Name, m.Value)
}
