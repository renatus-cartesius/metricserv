package config

import (
	"flag"
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
	if envSrvAddress := os.Getenv("ADDRESS"); envSrvAddress != "" {
		config.SrvAddress = envSrvAddress
	}

	flag.IntVar(&config.ReportInterval, "r", 10, "interval for reporting metrics to server")
	if envReportInterval := os.Getenv("REPORT_INTERVAL"); envReportInterval != "" {
		interval, err := strconv.ParseInt(envReportInterval, 10, 32)
		if err != nil {
			panic(err)
		}
		config.ReportInterval = int(interval)
	}

	flag.IntVar(&config.PollInterval, "p", 2, "interval for polling to server")
	if envPollInterval := os.Getenv("POLL_INTERVAL"); envPollInterval != "" {
		interval, err := strconv.ParseInt(envPollInterval, 10, 32)
		if err != nil {
			panic(err)
		}
		config.PollInterval = int(interval)
	}

	flag.StringVar(&config.AgentLogLevel, "l", "INFO", "logging level")
	if envAgentLogInterval := os.Getenv("AGENT_LOG_LEVEL"); envAgentLogInterval != "" {
		config.AgentLogLevel = envAgentLogInterval
	}
	flag.Parse()

	return config, nil
}
