package main

import (
	"flag"

	"github.com/renatus-cartesius/metricserv/internal/agent"
	"github.com/renatus-cartesius/metricserv/internal/monitor"
)

func main() {

	srvAddress := flag.String("a", "localhost:8080", "address to metrics server")
	reportInterval := flag.Int("r", 10, "interval for reporting metrics to server")
	pollInterval := flag.Int("p", 2, "interval for polling to server")
	flag.Parse()

	agent := agent.NewAgent(*reportInterval, *pollInterval, "http://"+*srvAddress, &monitor.MemMonitor{})
	agent.Serve()
}
