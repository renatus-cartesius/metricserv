package metrics

import "fmt"

type CounterMetric struct {
	Name  string `json:"name"`
	Value int64  `json:"value"`
	Type  string `json:"type"`
}

func NewCounter(name string, value int64) *CounterMetric {
	return &CounterMetric{
		Name:  name,
		Value: value,
		Type:  "counter",
	}
}

func (m CounterMetric) String() string {
	return fmt.Sprintf("counter:%s:%d", m.Name, m.Value)
}

func (m *CounterMetric) Change(value interface{}) error {
	m.Value += value.(int64)
	return nil
}

func (m *CounterMetric) GetValue() string {
	return fmt.Sprintf("%d", m.Value)
}

func (m *CounterMetric) GetType() string {
	return TypeCounter
}
