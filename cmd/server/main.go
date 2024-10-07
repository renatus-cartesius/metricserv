package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"

	"github.com/renatus-cartesius/metricserv/internal/server/handlers"
	"github.com/renatus-cartesius/metricserv/internal/storage"
)

func main() {
	srvAddress := flag.String("a", "localhost:8080", "address to metrics server")
	if envSrvAddress := os.Getenv("ADDRESS"); envSrvAddress != "" {
		*srvAddress = envSrvAddress
	}
	flag.Parse()

	memStorage := storage.NewMemStorage()
	srv := handlers.NewServerHandler(memStorage)

	r := chi.NewRouter()

	r.Route("/", func(r chi.Router) {
		r.Get("/", srv.AllMetrics)
		r.Get("/value/{type}/{name}", srv.GetValue)
		r.Post("/update/{type}/{name}/{value}", srv.Update)
	})

	// r.Post("/update/{type}/{name}/{value}", srv.Update)

	log.Fatalln(http.ListenAndServe(*srvAddress, r))
}
