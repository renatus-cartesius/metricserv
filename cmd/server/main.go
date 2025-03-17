package main

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"github.com/renatus-cartesius/metricserv/pkg/encryption"
	"github.com/renatus-cartesius/metricserv/pkg/proto"
	"github.com/renatus-cartesius/metricserv/pkg/server/pb"
	"github.com/renatus-cartesius/metricserv/pkg/utils"
	"google.golang.org/grpc"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/renatus-cartesius/metricserv/cmd/helpers"
	"github.com/renatus-cartesius/metricserv/cmd/server/config"
	"github.com/renatus-cartesius/metricserv/pkg/logger"
	"github.com/renatus-cartesius/metricserv/pkg/server/handlers"
	"github.com/renatus-cartesius/metricserv/pkg/storage"
	"go.uber.org/zap"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

var (
	buildDate    string
	buildCommit  string
	buildVersion string
)

func main() {

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

	logger.Log.Info(fmt.Sprintf("Build version: %v", utils.TagHelper(buildVersion)))
	logger.Log.Info(fmt.Sprintf("Build date: %v", utils.TagHelper(buildDate)))
	logger.Log.Info(fmt.Sprintf("Build commit: %v", utils.TagHelper(buildCommit)))

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
			logger.Log.Fatal(
				"error when restoring storage",
				zap.Error(err),
			)
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

	rsaProcessor, err := encryption.NewRSAProcessor()
	if err != nil {
		log.Fatalln(err)
	}

	privateKey, err := encryption.NewRSAPrivateKey(cfg.PrivateKey)
	if err != nil {
		log.Fatalln(err)
	}

	rsaProcessor.SetPrivateKey(privateKey)

	var trustedSubnet *net.IPNet
	if cfg.TrustedSubnet != "" {
		_, trustedSubnet, err = net.ParseCIDR(cfg.TrustedSubnet)
		if err != nil {
			log.Fatalln(err)
		}
	}

	srv := handlers.NewServerHandler(s, rsaProcessor, trustedSubnet)

	r := chi.NewRouter()

	server := &http.Server{Addr: cfg.SrvAddress, Handler: r}

	handlers.Setup(r, srv, cfg.HashKey)

	serverCtx, serverStopCtx := context.WithCancel(context.Background())
	defer serverStopCtx()

	shutdownSig := make(chan os.Signal, 1)
	signal.Notify(shutdownSig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// GRPC server setup
	listen, err := net.Listen("tcp", ":3200")
	if err != nil {
		logger.Log.Error(
			"error on creating grpc listen",
			zap.Error(err),
		)
	}

	wg := sync.WaitGroup{}
	gs := grpc.NewServer()
	proto.RegisterMetricsServiceServer(gs, &pb.Server{
		TrustedSubnet: trustedSubnet,
		Storage:       s,
		EncProcessor:  rsaProcessor,
	})

	wg.Add(1)
	go func() {
		logger.Log.Info("starting grpc server")
		if err := gs.Serve(listen); err != nil {
			logger.Log.Error(
				"error on listening grpc server",
				zap.Error(err),
			)
		}

		logger.Log.Info("shutting down grpc server")
		wg.Done()
	}()

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

		gs.GracefulStop()

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

	wg.Wait()
}
