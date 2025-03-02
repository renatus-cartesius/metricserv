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
	SrvAddress     string `flag:"-a"`
	SaveInterval   int    `flag:"-i"`
	RestoreStorage bool   `flag:"-r"`
	ServerLogLevel string `flag:"-l"`
	SavePath       string `flag:"-f"`
	DBDsn          string `flag:"-d"`
	HashKey        string `flag:"-k"`
	PrivateKey     string `flag:"-crypto-key"`
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

	flag.StringVar(&config.SrvAddress, "a", config.SrvAddress, "address to metrics server")
	flag.IntVar(&config.SaveInterval, "i", config.SaveInterval, "interval to storage file save")
	flag.BoolVar(&config.RestoreStorage, "r", defaults.RestoreStorage, "if true restoring server from file")
	flag.StringVar(&config.ServerLogLevel, "l", defaults.ServerLogLevel, "logging level")
	flag.StringVar(&config.SavePath, "f", defaults.SavePath, "path to storage file save")
	flag.StringVar(&config.DBDsn, "d", defaults.DBDsn, "connection string to database")
	flag.StringVar(&config.HashKey, "k", defaults.HashKey, "key for hashing payload")
	flag.StringVar(&config.PrivateKey, defaults.PrivateKey, "./private.pem", "private key")
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

	return config, nil
}
