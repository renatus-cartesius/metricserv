package metrics

import "fmt"

type CounterMetric struct {
	ID    string `json:"id"`
	Value int64  `json:"value"`
	Type  string `json:"type"`
}

func NewCounter(id string, value int64) *CounterMetric {
	return &CounterMetric{
		ID:    id,
		Value: value,
		Type:  TypeCounter,
	}
}

func (m CounterMetric) GetID() string {
	return m.ID
}

func (m *CounterMetric) String() string {
	return fmt.Sprintf("%s:%s:%d", TypeCounter, m.ID, m.Value)
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
