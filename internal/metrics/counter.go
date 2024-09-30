package metrics

import "fmt"

type CounterMetric struct {
	Name  string
	Value int64
}

func (m CounterMetric) String() string {
	return fmt.Sprintf("counter:%s:%d", m.Name, m.Value)
}

func (m *CounterMetric) Change(value interface{}) error {
	m.Value += value.(int64)
	return nil
}
