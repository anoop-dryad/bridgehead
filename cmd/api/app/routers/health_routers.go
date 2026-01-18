package routers

import (
	"github.com/anoop-dryad/bridgehead/cmd/api/app/handlers"
	"github.com/gin-gonic/gin"
)

func Health(router *gin.RouterGroup) {
	h := handlers.NewHealthHandler()

	router.GET("ping/", h.HealthCheck)
}
