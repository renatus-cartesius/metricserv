package config

import (
	"flag"
	"os"
	"strconv"
)

type Config struct {
	SrvAddress     string
	SaveInterval   int
	RestoreStorage bool
	ServerLogLevel string
	SavePath       string
}

func LoadConfig() (*Config, error) {

	config := &Config{}

	flag.StringVar(&config.SrvAddress, "a", "localhost:8080", "address to metrics server")
	if envSrvAddress := os.Getenv("ADDRESS"); envSrvAddress != "" {
		config.SrvAddress = envSrvAddress
	}

	flag.StringVar(&config.ServerLogLevel, "l", "INFO", "logging level")
	if envServerLogInterval := os.Getenv("SERVER_LOG_LEVEL"); envServerLogInterval != "" {
		config.ServerLogLevel = envServerLogInterval
	}

	flag.StringVar(&config.SavePath, "f", "./storage.json", "path to storage file save")
	if envSavePath := os.Getenv("FILE_STORAGE_PATH"); envSavePath != "" {
		config.SavePath = envSavePath
	}

	flag.IntVar(&config.SaveInterval, "i", 300, "interval to storage file save")
	if envSaveInterval := os.Getenv("STORE_INTERVAL"); envSaveInterval != "" {
		config.SaveInterval, _ = strconv.Atoi(envSaveInterval)
	}

	flag.BoolVar(&config.RestoreStorage, "r", true, "if true restoring server from file")
	if envRestoreStorage := os.Getenv("RESTORE"); envRestoreStorage != "" {
		config.RestoreStorage, _ = strconv.ParseBool(envRestoreStorage)
	}
	flag.Parse()

	return config, nil
}
