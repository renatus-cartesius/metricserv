package main

import (
	"net/http"

	"github.com/renatus-cartesius/metricserv/internal/server/handlers"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/update/{type}/{name}/{value}", handlers.UpdateHandler)

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)

	}
	// a := "123.123"
	// value, err := strconv.ParseFloat(a, 64)
	// if err != nil {
	// 	panic(err)
	// }

	// fmt.Println(int64(value))
	// fmt.Println(math.Round(value))
}
