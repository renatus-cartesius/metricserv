package main

import (
	"context"
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

	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	config, err := config.LoadConfig()
	if err != nil {
		log.Fatalln(err)
	}

	if err := logger.Initialize(config.AgentLogLevel); err != nil {
		log.Fatalln(err)
	}

	agent, err := agent.NewAgent(config.ReportInterval, config.PollInterval, "http://"+config.SrvAddress, &monitor.MemMonitor{}, config.HashKey)
	if err != nil {
		log.Fatal(err)
	}

	agent.Serve(ctx, config.RateLimit)
}
