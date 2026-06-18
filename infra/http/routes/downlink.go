package routes

import (
	"github.com/anoop-dryad/bridgehead/infra/http/handlers"
	"github.com/gin-gonic/gin"
)

func Downlink(router *gin.RouterGroup, h *handlers.DownlinkHandler) {
	router.POST("/downlinks", h.Create)
}
