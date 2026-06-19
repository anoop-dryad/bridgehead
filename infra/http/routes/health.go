package routes

import (
	"github.com/anoop-dryad/bridgehead/infra/http/handlers"
	"github.com/gin-gonic/gin"
)

func Health(router *gin.RouterGroup) {
	router.GET("ping/", handlers.HealthCheck)
}
