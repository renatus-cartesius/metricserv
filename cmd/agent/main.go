package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/renatus-cartesius/metricserv/cmd/agent/config"
	"github.com/renatus-cartesius/metricserv/internal/agent"
	"github.com/renatus-cartesius/metricserv/internal/logger"
	"github.com/renatus-cartesius/metricserv/internal/monitor"
)

func main() {

	exitCh := make(chan os.Signal, 1)
	signal.Notify(exitCh, os.Interrupt, syscall.SIGTERM)

	config, err := config.LoadConfig()
	if err != nil {
		log.Fatalln(err)
	}

	if err := logger.Initialize(config.AgentLogLevel); err != nil {
		log.Fatalln(err)
	}

	agent := agent.NewAgent(config.ReportInterval, config.PollInterval, "http://"+config.SrvAddress, &monitor.MemMonitor{}, exitCh, config.HashKey)

	agent.Serve(config.RateLimit)
}
