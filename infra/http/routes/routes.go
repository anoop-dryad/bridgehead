package routes

import (
	"github.com/anoop-dryad/bridgehead/config"
	"github.com/anoop-dryad/bridgehead/infra/http/handlers"
	"github.com/anoop-dryad/bridgehead/infra/http/swagger"
	"github.com/gin-gonic/gin"
)

func Register(engine *gin.Engine, deps handlers.Dependencies, cfg config.HTTP) {
	v1 := engine.Group("/v1")
	// v1.Use(middleware.Auth())

	Health(v1)
	Downlink(v1, deps.DownlinkHandler)

	if !cfg.IsProduction {
		swagger.Register(engine.Group("/"))
	}
}
