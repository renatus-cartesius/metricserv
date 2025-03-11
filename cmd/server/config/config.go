// Package config contains all configurable parameters for run metrics server
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

type Config struct {
	SrvAddress     string
	SaveInterval   int
	RestoreStorage bool
	ServerLogLevel string
	SavePath       string
	DBDsn          string
	HashKey        string
	PrivateKey     string
	TrustedSubnet  string
}

func LoadConfig() (*Config, error) {

	var err error
	var ok bool
	config := &Config{}
	defaults := &Config{
		SrvAddress:     "localhost:8080",
		ServerLogLevel: "DEBUG",
		SavePath:       "./storage.json",
		DBDsn:          "",
		SaveInterval:   300,
		RestoreStorage: true,
		HashKey:        "",
		PrivateKey:     "./private.pem",
		TrustedSubnet:  "",
	}

	configPath := "./server.json"
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

	flag.StringVar(&config.SrvAddress, "a", defaults.SrvAddress, "address to metrics server")
	flag.IntVar(&config.SaveInterval, "i", defaults.SaveInterval, "interval to storage file save")
	flag.BoolVar(&config.RestoreStorage, "r", defaults.RestoreStorage, "if true restoring server from file")
	flag.StringVar(&config.ServerLogLevel, "l", defaults.ServerLogLevel, "logging level")
	flag.StringVar(&config.SavePath, "f", defaults.SavePath, "path to storage file save")
	flag.StringVar(&config.DBDsn, "d", defaults.DBDsn, "connection string to database")
	flag.StringVar(&config.HashKey, "k", defaults.HashKey, "key for hashing payload")
	flag.StringVar(&config.PrivateKey, "p", defaults.PrivateKey, "private key")
	flag.StringVar(&config.TrustedSubnet, "t", defaults.TrustedSubnet, "agents trusted subnet")
	flag.StringVar(&configPath, "config", "./server.json", "path to config file")

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
	if envPrivateKey := os.Getenv("CRYPTO_KEY"); envPrivateKey != "" {
		config.PrivateKey = envPrivateKey
	}
	if envTrustedSubnet := os.Getenv("TRUSTED_SUBNET"); envTrustedSubnet != "" {
		config.TrustedSubnet = envTrustedSubnet
	}

	return config, nil
}
