package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type DownlinkHandler struct {
}

func NewDownlinkHandler() *DownlinkHandler {
	return &DownlinkHandler{}
}

func (h *DownlinkHandler) Create(c *gin.Context) {
	// var req dto.CreateDownlinkRequest
	// if err := c.ShouldBindJSON(&req); err != nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	// 	return
	// }
	// result, err := h.svc.Create(c.Request.Context(), req.ToModel())
	// if err != nil {
	// 	// error mapping lives here, not in domain
	// 	c.JSON(mapError(err), gin.H{"error": err.Error()})
	// 	return
	// }
	// c.JSON(http.StatusCreated, dto.FromDownlink(result))
	c.JSON(http.StatusCreated, gin.H{"msg": "create api being hit"})
}
