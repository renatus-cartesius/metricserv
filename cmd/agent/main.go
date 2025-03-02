package main

import (
	"context"
	"fmt"
	"github.com/renatus-cartesius/metricserv/cmd/helpers"
	"github.com/renatus-cartesius/metricserv/pkg/encryption"
	"github.com/renatus-cartesius/metricserv/pkg/utils"
	"log"
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

	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

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

	logger.Log.Info(fmt.Sprintf("Build version: %v", utils.TagHelper(buildVersion)))
	logger.Log.Info(fmt.Sprintf("Build date: %v", utils.TagHelper(buildDate)))
	logger.Log.Info(fmt.Sprintf("Build commit: %v", utils.TagHelper(buildCommit)))

	rsaProcessor, err := encryption.NewRSAProcessor()
	if err != nil {
		log.Fatalln(err)
	}

	publicKey, err := encryption.NewRSAPublicKey(config.PublicKey)
	if err != nil {
		log.Fatalln(err)
	}

	rsaProcessor.SetPublicKey(publicKey)

	agent, err := agent.NewAgent(config.ReportInterval, config.PollInterval, "http://"+config.SrvAddress, &monitor.MemMonitor{}, config.HashKey, rsaProcessor)
	if err != nil {
		log.Fatal(err)
	}

	agent.Serve(ctx, config.RateLimit)
}
