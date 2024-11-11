package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"

	"github.com/renatus-cartesius/metricserv/cmd/server/config"
	"github.com/renatus-cartesius/metricserv/internal/logger"
	"github.com/renatus-cartesius/metricserv/internal/server/handlers"
	"github.com/renatus-cartesius/metricserv/internal/server/middlewares"
	"github.com/renatus-cartesius/metricserv/internal/storage"
)

func main() {

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalln(err)
	}

	if err := logger.Initialize(cfg.ServerLogLevel); err != nil {
		log.Fatalln(err)
	}

	memStorage, err := storage.NewMemStorage(cfg.SavePath)
	if err != nil {
		log.Fatalln("error on creating memory storage")
	}

	pgStorage, err := storage.NewPGStorage(cfg.DBDsn)
	if err != nil {
		log.Fatalln("error on creating postgresql storage")
	}

	if cfg.RestoreStorage {
		if err := memStorage.Load(); err != nil {
			log.Fatalln(err)
		}
	}

	saveSig := make(chan os.Signal, 1)
	signal.Notify(saveSig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	if cfg.SaveInterval > 0 {

		saveTicker := time.NewTicker(time.Duration(cfg.SaveInterval) * time.Second)
		defer saveTicker.Stop()

		go func() {
			for {
				select {
				case <-saveSig:
					return
				case <-saveTicker.C:
					if err := memStorage.Save(); err != nil {
						logger.Log.Error(
							"error on saving storage",
							zap.Error(err),
						)
					}
				}
			}

		}()
	}

	srv := handlers.NewServerHandler(memStorage)

	r := chi.NewRouter()
	server := &http.Server{Addr: cfg.SrvAddress, Handler: r}

	r.Get("/ping", middlewares.Gzipper(logger.RequestLogger(func(w http.ResponseWriter, r *http.Request) {
		if ok := pgStorage.Ping(); !ok {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})))

	handlers.Setup(r, srv)

	serverCtx, serverStopCtx := context.WithCancel(context.Background())
	defer serverStopCtx()

	shutdownSig := make(chan os.Signal, 1)
	signal.Notify(shutdownSig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		<-shutdownSig

		logger.Log.Info(
			"graceful shuting down",
			zap.String("address", cfg.SrvAddress),
		)

		shutdownCtx, shutdownCancel := context.WithTimeout(serverCtx, 30*time.Second)
		defer shutdownCancel()

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				logger.Log.Fatal(
					"graceful shutdown timed out",
					zap.String("address", cfg.SrvAddress),
				)
			}
		}()

		err := server.Shutdown(shutdownCtx)
		if err != nil {
			logger.Log.Fatal(
				"error on graceful shutdown",
				zap.String("address", cfg.SrvAddress),
			)
		}

	}()

	logger.Log.Info(
		"starting server",
		zap.String("address", cfg.SrvAddress),
	)

	err = server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalln(err)
	}

	if err = memStorage.Save(); err != nil {
		log.Fatalln(err)
	}
}
