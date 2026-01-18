package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
}

func NewHealthHandler() HealthHandler {
	return HealthHandler{}
}

// @BasePath /api/v1

// PingExample godoc
// @Summary Use me to check if the service is healthy
// @Schemes
// @Description Ping me: and expect a response with Pong
// @Tags HealthCheck for the service
// @Accept json
// @Produce json
// @Success 200 {string} Pong
// @Router /ping [get]
func (hh HealthHandler) HealthCheck(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}
