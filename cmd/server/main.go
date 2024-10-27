package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/renatus-cartesius/metricserv/internal/logger"
	"github.com/renatus-cartesius/metricserv/internal/server/handlers"
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
	flag.Parse()

	logger.Initialize(*serverLogLevel)

	memStorage := storage.NewMemStorage()
	srv := handlers.NewServerHandler(memStorage)

	r := chi.NewRouter()

	r.Route("/", func(r chi.Router) {
		r.Get("/", logger.RequestLogger(srv.AllMetrics))
		r.Route("/value", func(r chi.Router) {
			r.Post("/", logger.RequestLogger(srv.GetValueJSON))
			r.Get("/{type}/{name}", logger.RequestLogger(srv.GetValue))
		})
		r.Route("/update", func(r chi.Router) {
			r.Post("/", logger.RequestLogger(srv.UpdateJSON))
			r.Post("/{type}/{name}/{value}", logger.RequestLogger(srv.Update))
		})
	})

	// r.Post("/update/{type}/{name}/{value}", srv.Update)

	logger.Log.Info(
		"starting server",
		zap.String("address", *srvAddress),
	)

	log.Fatalln(http.ListenAndServe(*srvAddress, r))
}
