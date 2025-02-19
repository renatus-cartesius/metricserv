package main

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"go.uber.org/zap"

	"github.com/renatus-cartesius/metricserv/cmd/helpers"
	"github.com/renatus-cartesius/metricserv/cmd/server/config"
	"github.com/renatus-cartesius/metricserv/pkg/logger"
	"github.com/renatus-cartesius/metricserv/pkg/server/handlers"
	"github.com/renatus-cartesius/metricserv/pkg/storage"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

var (
	buildDate    string
	buildCommit  string
	buildVersion string
)

func main() {

	fmt.Println("Build version:", tagHelper(buildVersion))
	fmt.Println("Build date:", tagHelper(buildDate))
	fmt.Println("Build commit:", tagHelper(buildCommit))

	ctx := context.Background()

	pprofCtx, pprofStopCtx := context.WithCancel(context.Background())
	defer pprofStopCtx()
	go helpers.SetupPprofHandlers(pprofCtx, ":8081")

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalln(err)
	}

	if err = logger.Initialize(cfg.ServerLogLevel); err != nil {
		log.Fatalln(err)
	}

	var s storage.Storager

	if cfg.DBDsn != "" {

		db, err := sql.Open("pgx", cfg.DBDsn)
		if err != nil {
			log.Fatalln(err)
		}

		goose.SetBaseFS(embedMigrations)

		if err = goose.SetDialect("postgres"); err != nil {
			logger.Log.Fatal(
				"error setting goose dialect",
				zap.Error(err),
			)
		}

		if err = goose.Up(db, "migrations"); err != nil {
			logger.Log.Fatal(
				"error on applying startup migration",
				zap.Error(err),
			)
		}

		s, err = storage.NewPGStorage(db)
		if err != nil {
			log.Fatalln("error on creating postgresql storage")
		}
		logger.Log.Info(
			"using postgresql as a storage backend",
		)
		defer func() {
			if err := db.Close(); err != nil {
				logger.Log.Fatal(
					"error on closing database",
					zap.Error(err),
				)
			}
		}()
	} else {
		s, err = storage.NewMemStorage(cfg.SavePath)
		if err != nil {
			log.Fatalln("error on creating memory storage")
		}
		logger.Log.Info(
			"using memorystorage as a storage backend",
		)
	}

	if cfg.RestoreStorage {
		if err = s.Load(ctx); err != nil {
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
					if err = s.Save(ctx); err != nil {
						logger.Log.Error(
							"error on saving storage",
							zap.Error(err),
						)
					}
				}
			}

		}()
	}

	srv := handlers.NewServerHandler(s)

	r := chi.NewRouter()

	server := &http.Server{Addr: cfg.SrvAddress, Handler: r}

	handlers.Setup(r, srv, cfg.HashKey)

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

		err = server.Shutdown(shutdownCtx)
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

	if err = s.Save(ctx); err != nil {
		log.Fatalln(err)
	}
}

func tagHelper(tag string) string {
	if tag == "" {
		return "N/A"
	} else {
		return tag
	}
}
