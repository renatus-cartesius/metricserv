package main

import (
	"github.com/renatus-cartesius/metricserv/internal/agent"
	"github.com/renatus-cartesius/metricserv/internal/monitor"
)

func main() {
	agent := agent.NewAgent(2, "http://localhost:8080", &monitor.MemMonitor{})
	agent.Serve()
}
