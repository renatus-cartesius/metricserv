// Package config contains all configurable parameters for run metrics server
package config

import (
	"flag"
	"log"
	"os"
	"strconv"
)

type Config struct {
	SrvAddress     string
	SaveInterval   int
	RestoreStorage bool
	ServerLogLevel string
	SavePath       string
	DBDsn          string
	HashKey        string
}

func LoadConfig() (*Config, error) {

	var err error
	config := &Config{}

	flag.StringVar(&config.SrvAddress, "a", "localhost:8080", "address to metrics server")
	flag.StringVar(&config.ServerLogLevel, "l", "INFO", "logging level")
	flag.StringVar(&config.SavePath, "f", "./storage.json", "path to storage file save")
	flag.StringVar(&config.DBDsn, "d", "", "connection string to database")
	flag.IntVar(&config.SaveInterval, "i", 300, "interval to storage file save")
	flag.BoolVar(&config.RestoreStorage, "r", true, "if true restoring server from file")
	flag.StringVar(&config.HashKey, "k", "", "key for hashing payload")

	flag.Parse()

	if envSrvAddress := os.Getenv("ADDRESS"); envSrvAddress != "" {
		config.SrvAddress = envSrvAddress
	}
	if envServerLogInterval := os.Getenv("SERVER_LOG_LEVEL"); envServerLogInterval != "" {
		config.ServerLogLevel = envServerLogInterval
	}
	if envSavePath := os.Getenv("FILE_STORAGE_PATH"); envSavePath != "" {
		config.SavePath = envSavePath
	}
	if envDBDsn := os.Getenv("DATABASE_DSN"); envDBDsn != "" {
		config.DBDsn = envDBDsn
	}
	if envSaveInterval := os.Getenv("STORE_INTERVAL"); envSaveInterval != "" {
		config.SaveInterval, err = strconv.Atoi(envSaveInterval)
		if err != nil {
			log.Fatal(err)
		}
	}
	if envRestoreStorage := os.Getenv("RESTORE"); envRestoreStorage != "" {
		config.RestoreStorage, err = strconv.ParseBool(envRestoreStorage)
		if err != nil {
			log.Fatal(err)
		}
	}
	if envHashKey := os.Getenv("KEY"); envHashKey != "" {
		config.HashKey = envHashKey
	}

	return config, nil
}
