package handlers

import (
	"net/http"

	"github.com/anoop-dryad/bridgehead/infra/http/dto"
	"github.com/anoop-dryad/bridgehead/internal/downlink"
	"github.com/gin-gonic/gin"
)

type DownlinkHandler struct {
	svc *downlink.Service
}

func NewDownlinkHandler(svc *downlink.Service) *DownlinkHandler {
	return &DownlinkHandler{svc: svc}
}

// Create godoc
// @Summary      Create downlink request
// @Tags         downlink
// @Accept       json
// @Produce      json
// @Param        request body dto.CreateDownlinkRequest true "request"
// @Success      201 {object} dto.DownlinkResponse
// @Failure      400 {object} dto.ErrorResponse
// @Failure      409 {object} dto.ErrorResponse
// @Router       /v1/downlinks [post]
func (h *DownlinkHandler) Create(c *gin.Context) {
	var req dto.CreateDownlinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}
	model, err := req.ToModel()
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid payload encoding"})
		return
	}
	result, err := h.svc.Create(c.Request.Context(), model)
	if err != nil {
		c.JSON(mapError(err), dto.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, dto.FromDownlink(result))
}

// Get godoc
// @Summary      Get downlink request
// @Tags         downlink
// @Produce      json
// @Param        id path string true "request id"
// @Success      200 {object} dto.DownlinkResponse
// @Failure      404 {object} dto.ErrorResponse
// @Router       /v1/downlinks/{id} [get]
func (h *DownlinkHandler) Get(c *gin.Context) {
	result, err := h.svc.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(mapError(err), dto.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.FromDownlink(result))
}

// Delete godoc
// @Summary      Delete downlink request
// @Tags         downlink
// @Param        id path string true "request id"
// @Success      204
// @Failure      404 {object} dto.ErrorResponse
// @Failure      409 {object} dto.ErrorResponse
// @Router       /v1/downlinks/{id} [delete]
func (h *DownlinkHandler) Delete(c *gin.Context) {
	if err := h.svc.Delete(c.Request.Context(), c.Param("id")); err != nil {
		c.JSON(mapError(err), dto.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

// List godoc
// @Summary      List downlink requests by device EUI
// @Tags         downlink
// @Produce      json
// @Param        device_eui query string true "device EUI"
// @Success      200 {array} dto.DownlinkResponse
// @Failure      400 {object} dto.ErrorResponse
// @Router       /v1/downlinks [get]
func (h *DownlinkHandler) List(c *gin.Context) {
	deviceEUI := c.Query("device_eui")
	if deviceEUI == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "device_eui is required"})
		return
	}
	results, err := h.svc.List(c.Request.Context(), deviceEUI)
	if err != nil {
		c.JSON(mapError(err), dto.ErrorResponse{Error: err.Error()})
		return
	}
	response := make([]dto.DownlinkResponse, len(results))
	for i, r := range results {
		response[i] = dto.FromDownlink(r)
	}
	c.JSON(http.StatusOK, response)
}
