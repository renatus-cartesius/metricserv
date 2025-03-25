// Package config contains all configurable parameters for run metrics agent
package config

import (
	"encoding/json"
	"flag"
	"github.com/renatus-cartesius/metricserv/pkg/logger"
	"log"
	"os"
	"strconv"
	"strings"
)

type AgentConfig struct {
	SrvAddress     string
	ReportInterval int
	PollInterval   int
	AgentLogLevel  string
	HashKey        string
	RateLimit      int
	PublicKey      string
}

func LoadAgentConfig() (*AgentConfig, error) {

	var err error
	var ok bool
	config := &AgentConfig{}
	defaults := &AgentConfig{
		SrvAddress:     "localhost:8080",
		ReportInterval: 10,
		PollInterval:   2,
		AgentLogLevel:  "INFO",
		HashKey:        "",
		RateLimit:      2,
		PublicKey:      "./public.pem",
	}

	configPath := "./agent.json"
	log.Println("getting path to config from CONFIG env")

	if configPath, ok = os.LookupEnv("CONFIG"); !ok {

		log.Println("env CONFIG is not set, checking if -config flag passed")
		for i, arg := range os.Args[:1] {
			if !strings.HasPrefix(arg, "-") {
				continue
			}

			if strings.Contains(arg, "=") {
				if strings.Split(arg, "=")[0] == "-config" {
					log.Println("config path found in flag")
					configPath = strings.Split(arg, "=")[1]
					break
				}
			} else {
				if arg == "-config" {
					log.Println("config path found in flag")
					configPath = os.Args[i+1]
					break
				}
			}
		}

	} else {
		log.Println("config path gotten from env CONFIG")
	}

	configFile, err := os.Open(configPath)
	if err != nil {
		log.Println("cannot open config file: ", err)
	}

	if err := json.NewDecoder(configFile).Decode(&defaults); err != nil {
		logger.Log.Error("cannot unmarshall config file content")
	}

	flag.StringVar(&config.SrvAddress, "a", "localhost:8080", "address to metrics server")
	flag.IntVar(&config.ReportInterval, "r", 10, "interval for reporting metrics to server")
	flag.IntVar(&config.PollInterval, "p", 2, "interval for polling to server")
	flag.StringVar(&config.AgentLogLevel, "l", "INFO", "logging level")
	flag.StringVar(&config.HashKey, "k", "", "key for hashing payload")
	flag.IntVar(&config.RateLimit, "L", 2, "amount of a parallel workers")
	flag.StringVar(&config.PublicKey, "crypto-key", "./public.pem", "public key")
	flag.StringVar(&configPath, "config", "./agent.json", "path to config file")

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
