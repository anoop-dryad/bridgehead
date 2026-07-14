package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/anoop-dryad/bridgehead/app/config"
	"github.com/anoop-dryad/bridgehead/app/infra/cache"
	"github.com/anoop-dryad/bridgehead/app/infra/db"
	"github.com/anoop-dryad/bridgehead/app/infra/http/handlers"
	"github.com/anoop-dryad/bridgehead/app/infra/http/server"
	"github.com/anoop-dryad/bridgehead/app/infra/kinesis"
	"github.com/anoop-dryad/bridgehead/app/infra/logger"
	gatewaymqtt "github.com/anoop-dryad/bridgehead/app/infra/mqtt/gateway"
	"github.com/anoop-dryad/bridgehead/app/infra/mqtt/ttn"
	"github.com/anoop-dryad/bridgehead/app/infra/sqs"
	"github.com/anoop-dryad/bridgehead/app/internal/downlink"
	"github.com/anoop-dryad/bridgehead/app/internal/gateway"
	"github.com/anoop-dryad/bridgehead/app/internal/routing"
	"github.com/anoop-dryad/bridgehead/app/internal/scheduler"
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
	redisCache := cache.NewRedisCache(cfg.Redis)
	if err := redisCache.Ping(ctx); err != nil {
		log.Fatal("redis unreachable", zap.Error(err))
	}
	defer redisCache.Close()

	// repos
	downlinkRepo := downlink.NewRepository(dbPool)
	sensorRepo := sensor.NewRepository(dbPool)
	gatewayRepo := gateway.NewRepository(dbPool)

	// services : logging is only supported here
	sensorSvc := sensor.NewService(sensorRepo, log)
	gatewaySvc := gateway.NewService(gatewayRepo, redisCache, log)
	resolver := routing.New(
		routing.NewSensorAdapter(sensorSvc),
		routing.NewGatewayAdapter(gatewaySvc),
	)
	downlinkSvc := downlink.NewService(downlinkRepo, resolver, log)

	// gateway publisher
	gatewayPub, err := gatewaymqtt.NewPublisher(cfg.MQTT.Gateway, log)
	if err != nil {
		log.Fatal("gateway publisher init failed", zap.Error(err))
	}
	defer gatewayPub.Disconnect()

	// ttn publisher
	ttnPub, err := ttn.NewPublisher(cfg.MQTT.TTN, log)
	if err != nil {
		log.Fatal("ttn publisher init failed", zap.Error(err))
	}
	defer ttnPub.Disconnect()

	// dispatcher
	dispatcher := scheduler.NewDispatcher(downlinkSvc, sensorSvc, gatewaySvc, gatewayPub, ttnPub, resolver, log)

	// gateway mqtt consumer
	gatewayConsumer, err := gatewaymqtt.NewConsumer(cfg.MQTT.Gateway, gatewaySvc, dispatcher, log)
	if err != nil {
		log.Fatal("failed to init gateway mqtt consumer", zap.Error(err))
	}
	go gatewayConsumer.Start(ctx)

	// kinesis consumer
	kinesisConsumer, err := kinesis.NewConsumer(cfg.Kinesis, sensorSvc, log)
	if err != nil {
		log.Fatal("failed to init kinesis consumer", zap.Error(err))
	}
	go kinesisConsumer.Start(ctx)

	// sqs consumer
	sqsConsumer, err := sqs.NewConsumer(cfg.SQS, sensorSvc, gatewaySvc, log)
	if err != nil {
		log.Fatal("sqs consumer init failed", zap.Error(err))
	}
	go sqsConsumer.Start(ctx)

	// downlink expiry watcher
	expiryWatcher := scheduler.NewExpiryWatcher(downlinkSvc, log)
	go expiryWatcher.Run(ctx)

	// handlers
	deps := handlers.Dependencies{
		DownlinkHandler: handlers.NewDownlinkHandler(downlinkSvc),
	}

	srv := server.NewServer(cfg.App, deps, log)
	log.Info("starting server", zap.String("addr", cfg.HTTP.Addr))

	if err := srv.Start(ctx); err != nil {
		log.Fatal("server failed", zap.Error(err))
	}

	log.Info("server stopped gracefully")

}
