package monitor

type Monitor interface {
	Get() map[string]string
	Flush() error
}

type RuntimeMetric struct {
	Name  string
	Value string
}
