package main

import (
	"github.com/anoop-dryad/bridgehead/config"
	"github.com/anoop-dryad/bridgehead/infra/http/handlers"
	"github.com/anoop-dryad/bridgehead/infra/http/server"
)

func main() {
	cfg := config.Load()
	// db := infra.NewPostgresPool(cfg.DB)
	// cache := infra.NewRedisCache(cfg.Cache)

	// repos
	// gatewayRepo := gateway.NewRepository(db)
	// sensorRepo := sensor.NewRepository(db)
	// downlinkRepo := downlink.NewRepository(db)

	// services
	// gatewayService := gateway.NewService(gatewayRepo, cache)
	// sensorService := sensor.NewService(sensorRepo, gatewayService)
	// downlinkService := downlink.NewService(downlinkRepo, sensorService, gatewayService)

	// handlers
	deps := handlers.Dependencies{
		DownlinkHandler: handlers.NewDownlinkHandler(),
	}

	srv := server.NewServer(cfg.HTTP, deps)
	srv.ListenAndServe()
}
