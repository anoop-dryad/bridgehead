// infra/http/handlers/errors.go
package handlers

import (
	"errors"
	"net/http"

	"github.com/anoop-dryad/bridgehead/internal/downlink"
)

func mapError(err error) int {
	switch {
	case errors.Is(err, downlink.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, downlink.ErrDuplicateID):
		return http.StatusConflict
	case errors.Is(err, downlink.ErrInvalidStatus):
		return http.StatusConflict
	case errors.Is(err, downlink.ErrInvalidDeviceType),
		errors.Is(err, downlink.ErrInvalidDownlinkType):
		return http.StatusBadRequest
	case errors.Is(err, downlink.ErrExpired):
		return http.StatusGone
	default:
		return http.StatusInternalServerError
	}
}
