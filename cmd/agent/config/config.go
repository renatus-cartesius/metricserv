// Package config contains all configurable parameters for run metrics agent
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
	HashKey        string
	RateLimit      int
	PublicKey      string
}

func LoadConfig() (*Config, error) {

	config := &Config{}

	flag.StringVar(&config.SrvAddress, "a", "localhost:8080", "address to metrics server")
	flag.IntVar(&config.ReportInterval, "r", 10, "interval for reporting metrics to server")
	flag.IntVar(&config.PollInterval, "p", 2, "interval for polling to server")
	flag.StringVar(&config.AgentLogLevel, "l", "INFO", "logging level")
	flag.StringVar(&config.HashKey, "k", "", "key for hashing payload")
	flag.IntVar(&config.RateLimit, "L", 2, "amount of a parallel workers")
	flag.StringVar(&config.PublicKey, "crypto-key", "./public.pem", "public key")

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
	if envHashKey := os.Getenv("KEY"); envHashKey != "" {
		config.HashKey = envHashKey
	}
	if envRateLimit := os.Getenv("RATE_LIMIT"); envRateLimit != "" {
		rateLimit, err := strconv.ParseInt(envRateLimit, 10, 32)
		if err != nil {
			log.Fatalln(err)
		}
		config.RateLimit = int(rateLimit)

	}
	if envPublicKey := os.Getenv("CRYPTO_KEY"); envPublicKey != "" {
		config.PublicKey = envPublicKey
	}

	return config, nil
}
