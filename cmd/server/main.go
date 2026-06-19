package main

import (
	"github.com/anoop-dryad/bridgehead/config"
	"github.com/anoop-dryad/bridgehead/infra/db"
	"github.com/anoop-dryad/bridgehead/infra/http/handlers"
	"github.com/anoop-dryad/bridgehead/infra/http/server"
	"github.com/anoop-dryad/bridgehead/internal/downlink"
)

func main() {
	cfg := config.Load()
	db := db.NewPostgresPool(cfg.DB)

	// repos
	downlinkRepo := downlink.NewRepository(db)

	// services
	downlinkService := downlink.NewService(downlinkRepo)

	// handlers
	deps := handlers.Dependencies{
		DownlinkHandler: handlers.NewDownlinkHandler(downlinkService),
	}

	srv := server.NewServer(cfg.HTTP, deps)
	srv.ListenAndServe()
}
