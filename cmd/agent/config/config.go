package config

import (
	"flag"
	"log"
	"os"
	"strconv"
)

type Config struct {
	SrvAddress     string
	ReportInterval int
	PollInterval   int
	AgentLogLevel  string
}

func LoadConfig() (*Config, error) {

	config := &Config{}

	flag.StringVar(&config.SrvAddress, "a", "localhost:8080", "address to metrics server")
	flag.IntVar(&config.ReportInterval, "r", 10, "interval for reporting metrics to server")
	flag.IntVar(&config.PollInterval, "p", 2, "interval for polling to server")
	flag.StringVar(&config.AgentLogLevel, "l", "INFO", "logging level")

	flag.Parse()

	if envSrvAddress := os.Getenv("ADDRESS"); envSrvAddress != "" {
		config.SrvAddress = envSrvAddress
	}
	if envReportInterval := os.Getenv("REPORT_INTERVAL"); envReportInterval != "" {
		interval, err := strconv.ParseInt(envReportInterval, 10, 32)
		if err != nil {
			log.Fatalln(err)
		}
		config.ReportInterval = int(interval)
	}
	if envPollInterval := os.Getenv("POLL_INTERVAL"); envPollInterval != "" {
		interval, err := strconv.ParseInt(envPollInterval, 10, 32)
		if err != nil {
			log.Fatalln(err)
		}
		config.PollInterval = int(interval)
	}
	if envAgentLogInterval := os.Getenv("AGENT_LOG_LEVEL"); envAgentLogInterval != "" {
		config.AgentLogLevel = envAgentLogInterval
	}

	return config, nil
}
