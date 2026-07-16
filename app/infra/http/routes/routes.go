package routes

import (
	"github.com/anoop-dryad/bridgehead/app/infra/http/handlers"
	"github.com/anoop-dryad/bridgehead/app/infra/http/swagger"
	"github.com/gin-gonic/gin"
)

func Register(engine *gin.Engine, deps handlers.Dependencies, isProduction bool) {
	v1 := engine.Group("/v1")
	// v1.Use(middleware.Auth())

	Health(v1)
	Downlink(v1, deps.DownlinkHandler)

	// future
	// v2 := engine.Group("/v2")
	// Downlink(v2, deps.DownlinkV2Handler)

	if !isProduction {
		swagger.Register(engine, "v1")
		// swagger.Register(engine, "v2")  // when v2 exists
	}
}
