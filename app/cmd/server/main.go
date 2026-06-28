package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/anoop-dryad/bridgehead/app/config"
	"github.com/anoop-dryad/bridgehead/app/infra/db"
	"github.com/anoop-dryad/bridgehead/app/infra/http/handlers"
	"github.com/anoop-dryad/bridgehead/app/infra/http/server"
	"github.com/anoop-dryad/bridgehead/app/infra/kinesis"
	"github.com/anoop-dryad/bridgehead/app/infra/logger"
	"github.com/anoop-dryad/bridgehead/app/internal/downlink"
	"github.com/anoop-dryad/bridgehead/app/internal/sensor"
	"go.uber.org/zap"
)

// @title           Bridgehead API
// @version         1.0
// @description     Downlink service for border gateways and sensors
// @host            localhost:8080
func main() {
	cfg := config.Load()

	log, err := logger.New(cfg.App)
	if err != nil {
		panic("failed to init logger: " + err.Error())
	}
	defer log.Sync() // flush buffer before exit

	// context cancelled on OS signal — drives graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	dbPool := db.NewPostgresPool(cfg.DB)

	// repos
	downlinkRepo := downlink.NewRepository(dbPool)
	sensorRepo := sensor.NewRepository(dbPool)

	// services
	downlinkService := downlink.NewService(downlinkRepo, log) // logging only supported in service layer
	sensorSvc := sensor.NewService(sensorRepo, log)

	kinesisConsumer, err := kinesis.NewConsumer(cfg.Kinesis, sensorSvc, log)
	if err != nil {
		log.Fatal("failed to init kinesis consumer", zap.Error(err))
	}

	// start as background goroutine
	go kinesisConsumer.Start(ctx)

	// handlers
	deps := handlers.Dependencies{
		DownlinkHandler: handlers.NewDownlinkHandler(downlinkService),
	}

	srv := server.NewServer(cfg.App, deps, log)
	log.Info("starting server", zap.String("addr", cfg.HTTP.Addr))

	if err := srv.Start(ctx); err != nil {
		log.Fatal("server failed", zap.Error(err))
	}

	log.Info("server stopped gracefully")

}
