package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/renatus-cartesius/metricserv/internal/logger"
	"github.com/renatus-cartesius/metricserv/internal/server/handlers"
	"github.com/renatus-cartesius/metricserv/internal/server/middlewares"
	"github.com/renatus-cartesius/metricserv/internal/storage"
)

func main() {
	srvAddress := flag.String("a", "localhost:8080", "address to metrics server")
	if envSrvAddress := os.Getenv("ADDRESS"); envSrvAddress != "" {
		*srvAddress = envSrvAddress
	}

	serverLogLevel := flag.String("l", "INFO", "logging level")
	if envServerLogInterval := os.Getenv("SERVER_LOG_LEVEL"); envServerLogInterval != "" {
		*serverLogLevel = envServerLogInterval
	}

	savePath := flag.String("f", "./storage.json", "path to storage file save")
	if envSavePath := os.Getenv("FILE_STORAGE_PATH"); envSavePath != "" {
		*savePath = envSavePath
	}

	saveInterval := flag.Int("i", 300, "interval to storage file save")
	if envSaveInterval := os.Getenv("STORE_INTERVAL"); envSaveInterval != "" {
		*saveInterval, _ = strconv.Atoi(envSaveInterval)
	}

	restoreStorage := flag.Bool("r", true, "if true restoring server from file")
	if envRestoreStorage := os.Getenv("RESTORE"); envRestoreStorage != "" {
		*restoreStorage, _ = strconv.ParseBool(envRestoreStorage)
	}

	flag.Parse()

	logger.Initialize(*serverLogLevel)

	// file, _ := os.OpenFile("./storage.json", os.O_RDONLY, 0666)
	// testStorage := &storage.MemStorage{}
	// if err := json.NewDecoder(file).Decode(testStorage); err != nil {
	// 	panic(err)
	// }

	// fmt.Println(testStorage)

	memStorage := storage.NewMemStorage(*savePath)

	if *restoreStorage {
		memStorage.Load()
	}

	if *saveInterval > 0 {

		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		saveTicker := time.NewTicker(time.Duration(*saveInterval) * time.Second)

		go func() {
			for {
				select {
				case <-sig:
					return
				case <-saveTicker.C:
					memStorage.Save()
				}
			}

		}()
	}

	srv := handlers.NewServerHandler(memStorage)

	r := chi.NewRouter()
	server := &http.Server{Addr: *srvAddress, Handler: r}

	r.Route("/", func(r chi.Router) {
		r.Get("/", middlewares.Gzipper(logger.RequestLogger(srv.AllMetrics)))
		r.Route("/value", func(r chi.Router) {
			r.Post("/", middlewares.Gzipper(logger.RequestLogger(srv.GetValueJSON)))
			r.Get("/{type}/{name}", middlewares.Gzipper(logger.RequestLogger(srv.GetValue)))
		})
		r.Route("/update", func(r chi.Router) {
			r.Post("/", middlewares.Gzipper(logger.RequestLogger(srv.UpdateJSON)))
			r.Post("/{type}/{name}/{value}", middlewares.Gzipper(logger.RequestLogger(srv.Update)))
		})
	})

	// r.Post("/update/{type}/{name}/{value}", srv.Update)

	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		<-sig

		logger.Log.Info(
			"graceful shuting down",
			zap.String("address", *srvAddress),
		)

		shutdownCtx, _ := context.WithTimeout(serverCtx, 30*time.Second)

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				logger.Log.Fatal(
					"graceful shutdown timed out",
					zap.String("address", *srvAddress),
				)
			}
		}()

		err := server.Shutdown(shutdownCtx)
		if err != nil {
			logger.Log.Fatal(
				"error on graceful shutdown",
				zap.String("address", *srvAddress),
			)
		}

		serverStopCtx()
	}()

	logger.Log.Info(
		"starting server",
		zap.String("address", *srvAddress),
	)

	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		panic(err)
	}

	// if err = memStorage.Save(); err != nil {
	// 	panic(err)
	// }
}
