package main

import (
	"net/http"

	"github.com/renatus-cartesius/metricserv/internal/server/handlers"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /update/{type}/{name}/{value}", handlers.UpdateHandler)

	err := http.ListenAndServe("localhost:8080", mux)
	if err != nil {
		panic(err)

	}
}
