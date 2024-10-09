package monitor

type Monitor interface {
	Get() map[string]string
	Flush() error
}
