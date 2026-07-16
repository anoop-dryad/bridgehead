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
	appLog := newLogger(cfg.App)
	defer appLog.Sync() // flush buffer before exit

	// context cancelled on OS signal — drives graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// ------------------------------- infrastructure ---------------------------------------

	dbPool := db.NewPostgresPool(cfg.DB)
	redisCache := newRedisCache(ctx, cfg.Redis, appLog)
	defer redisCache.Close()

	// ------------------------------- domain ---------------------------------------

	sensorSvc := sensor.NewService(sensor.NewRepository(dbPool), appLog)
	gatewaySvc := gateway.NewService(gateway.NewRepository(dbPool), redisCache, appLog)
	resolver := routing.New(
		routing.NewSensorAdapter(sensorSvc),
		routing.NewGatewayAdapter(gatewaySvc),
	)
	downlinkSvc := downlink.NewService(downlink.NewRepository(dbPool), resolver, appLog)

	// ------------------------------- publishers ---------------------------------------

	gatewayPublisher := newGatewayMQTTPublisher(cfg.MQTT.Gateway, appLog)
	defer gatewayPublisher.Disconnect()
	ttnPublisher := newTTNPublisher(cfg.MQTT.TTN, appLog)
	defer ttnPublisher.Disconnect()

	// ------------------------------- dispatcher ---------------------------------------

	// TODO : need to revist the dependencies.
	dispatcher := scheduler.NewDispatcher(downlinkSvc, sensorSvc, gatewaySvc, gatewayPublisher, ttnPublisher, resolver, appLog)

	// ------------------------------- background workers ---------------------------------------

	go newGatewayMQTTConsumer(cfg.MQTT.Gateway, gatewaySvc, dispatcher, appLog).Start(ctx)
	go newKinesisConsumer(cfg.Kinesis, sensorSvc, appLog).Start(ctx)
	go newSqsConsumer(cfg.SQS, sensorSvc, gatewaySvc, appLog).Start(ctx)
	go scheduler.NewExpiryWatcher(downlinkSvc, appLog).Run(ctx)

	// ------------------------------- http server (blocks) ---------------------------------------

	deps := handlers.Dependencies{
		DownlinkHandler: handlers.NewDownlinkHandler(downlinkSvc),
	}
	srv := server.NewServer(cfg.HTTP, cfg.App, deps, appLog)
	appLog.Info("starting server", zap.String("addr", cfg.HTTP.Addr))
	if err := srv.Start(ctx); err != nil {
		appLog.Fatal("server failed", zap.Error(err))
	}
	appLog.Info("server stopped gracefully")

}

// ------------------------------- helper functions ---------------------------------------

func newLogger(appConfig config.App) *zap.Logger {
	appLog, err := logger.New(appConfig)
	if err != nil {
		panic("failed to init logger: " + err.Error())
	}

	return appLog
}

func newRedisCache(ctx context.Context, redisConfig config.Redis, appLog *zap.Logger) *cache.RedisCache {
	redisCache := cache.NewRedisCache(redisConfig)
	if err := redisCache.Ping(ctx); err != nil {
		appLog.Fatal("redis unreachable", zap.Error(err))
	}

	return redisCache
}

func newGatewayMQTTPublisher(mqttConfig config.GatewayMQTT, appLog *zap.Logger) *gatewaymqtt.Publisher {
	gatewayPub, err := gatewaymqtt.NewPublisher(mqttConfig, appLog)
	if err != nil {
		appLog.Fatal("gateway publisher init failed", zap.Error(err))
	}
	return gatewayPub
}

func newTTNPublisher(mqttConfig config.TTNMQTT, appLog *zap.Logger) *ttn.Publisher {
	ttnPub, err := ttn.NewPublisher(mqttConfig, appLog)
	if err != nil {
		appLog.Fatal("ttn publisher init failed", zap.Error(err))
	}
	return ttnPub
}

func newGatewayMQTTConsumer(mqttConfig config.GatewayMQTT, gatewaySvc *gateway.Service, dispatcher *scheduler.Dispatcher, appLog *zap.Logger) *gatewaymqtt.Consumer {
	gatewayConsumer, err := gatewaymqtt.NewConsumer(mqttConfig, gatewaySvc, dispatcher, appLog)
	if err != nil {
		appLog.Fatal("failed to init gateway mqtt consumer", zap.Error(err))
	}
	return gatewayConsumer
}

func newKinesisConsumer(kinesisConfig config.Kinesis, sensorSvc *sensor.Service, appLog *zap.Logger) *kinesis.Consumer {
	kinesisConsumer, err := kinesis.NewConsumer(kinesisConfig, sensorSvc, appLog)
	if err != nil {
		appLog.Fatal("failed to init kinesis consumer", zap.Error(err))
	}
	return kinesisConsumer
}

func newSqsConsumer(sqsConfig config.SQS, sensorSvc *sensor.Service, gatewaySvc *gateway.Service, appLog *zap.Logger) *sqs.Consumer {
	sqsConsumer, err := sqs.NewConsumer(sqsConfig, sensorSvc, gatewaySvc, appLog)
	if err != nil {
		appLog.Fatal("sqs consumer init failed", zap.Error(err))
	}

	return sqsConsumer
}
