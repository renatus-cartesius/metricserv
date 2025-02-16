// Package helpers contains of some useful helping functions
package helpers

import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/renatus-cartesius/metricserv/pkg/logger"
	"go.uber.org/zap"
	"log"
	"net/http"
	"net/http/pprof"
)

func SetupPprofHandlers(ctx context.Context, addr string) {
	r := chi.NewRouter()

	r.HandleFunc("/debug/pprof/", pprof.Index)
	r.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	r.HandleFunc("/debug/pprof/profile", pprof.Profile)
	r.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	r.HandleFunc("/debug/pprof/trace", pprof.Trace)

	logger.Log.Info(
		"starting pprof server",
		zap.String("address", addr),
	)
	pprofServer := &http.Server{Addr: addr, Handler: r}
	go func() {
		err := pprofServer.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalln(err)
		}
	}()
	<-ctx.Done()
	logger.Log.Info(
		"shutting down pprof server",
		zap.String("address", addr),
	)
	if err := pprofServer.Shutdown(ctx); err != nil {
		logger.Log.Error(
			"error on shutting down pprof server",
			zap.Error(err),
		)
	}

}
