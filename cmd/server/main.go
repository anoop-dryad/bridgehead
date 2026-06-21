package main

import (
	"github.com/anoop-dryad/bridgehead/config"
	"github.com/anoop-dryad/bridgehead/infra/db"
	"github.com/anoop-dryad/bridgehead/infra/http/handlers"
	"github.com/anoop-dryad/bridgehead/infra/http/server"
	"github.com/anoop-dryad/bridgehead/infra/logger"
	"github.com/anoop-dryad/bridgehead/internal/downlink"
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

	db := db.NewPostgresPool(cfg.DB)

	// repos
	downlinkRepo := downlink.NewRepository(db)

	// services
	downlinkService := downlink.NewService(downlinkRepo, log) // logging only supported in service layer

	// handlers
	deps := handlers.Dependencies{
		DownlinkHandler: handlers.NewDownlinkHandler(downlinkService),
	}

	srv := server.NewServer(cfg.App, deps, log)
	log.Info("starting server", zap.String("addr", cfg.HTTP.Addr))

	if err := srv.ListenAndServe(); err != nil {
		log.Fatal("server failed", zap.Error(err))
	}
}
