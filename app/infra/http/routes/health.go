package routes

import (
	"github.com/anoop-dryad/bridgehead/app/infra/http/handlers"
	"github.com/gin-gonic/gin"
)

func Health(router *gin.RouterGroup) {
	router.GET("ping/", handlers.HealthCheck)
}
