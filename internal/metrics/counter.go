package metrics

import "fmt"

type CounterMetric struct {
	Name  string
	Value int64
	Type  string
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
