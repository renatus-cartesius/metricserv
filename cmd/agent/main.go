package main

import (
	"flag"
	"os"
	"strconv"

	"github.com/renatus-cartesius/metricserv/internal/agent"
	"github.com/renatus-cartesius/metricserv/internal/monitor"
)

func main() {

	srvAddress := flag.String("a", "localhost:8080", "address to metrics server")
	if envSrvAddress := os.Getenv("ADDRESS"); envSrvAddress != "" {
		*srvAddress = envSrvAddress
	}

	reportInterval := flag.Int("r", 10, "interval for reporting metrics to server")
	if envReportInterval := os.Getenv("REPORT_INTERVAL"); envReportInterval != "" {
		interval, err := strconv.ParseInt(envReportInterval, 10, 32)
		if err != nil {
			panic(err)
		}
		*reportInterval = int(interval)
	}

	pollInterval := flag.Int("p", 2, "interval for polling to server")
	if envPollInterval := os.Getenv("POLL_INTERVAL"); envPollInterval != "" {
		interval, err := strconv.ParseInt(envPollInterval, 10, 32)
		if err != nil {
			panic(err)
		}
		*pollInterval = int(interval)
	}
	flag.Parse()

	agent := agent.NewAgent(*reportInterval, *pollInterval, "http://"+*srvAddress, &monitor.MemMonitor{})
	agent.Serve()
}
