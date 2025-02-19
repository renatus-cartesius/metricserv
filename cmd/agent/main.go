package main

import (
	"context"
	"fmt"
	"github.com/renatus-cartesius/metricserv/cmd/helpers"
	"github.com/renatus-cartesius/metricserv/pkg/utils"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/renatus-cartesius/metricserv/cmd/agent/config"
	"github.com/renatus-cartesius/metricserv/pkg/agent"
	"github.com/renatus-cartesius/metricserv/pkg/logger"
	"github.com/renatus-cartesius/metricserv/pkg/monitor"
)

var (
	buildDate    string
	buildCommit  string
	buildVersion string
)

func main() {

	fmt.Println("Build version:", utils.TagHelper(buildVersion))
	fmt.Println("Build date:", utils.TagHelper(buildDate))
	fmt.Println("Build commit:", utils.TagHelper(buildCommit))

	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	pprofCtx, pprofStopCtx := context.WithCancel(context.Background())
	defer pprofStopCtx()
	go helpers.SetupPprofHandlers(pprofCtx, ":8082")

	config, err := config.LoadConfig()
	if err != nil {
		log.Fatalln(err)
	}

	if err = logger.Initialize(config.AgentLogLevel); err != nil {
		log.Fatalln(err)
	}

	agent, err := agent.NewAgent(config.ReportInterval, config.PollInterval, "http://"+config.SrvAddress, &monitor.MemMonitor{}, config.HashKey)
	if err != nil {
		log.Fatal(err)
	}

	agent.Serve(ctx, config.RateLimit)
}
