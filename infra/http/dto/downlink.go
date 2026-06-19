// infra/http/dto/downlink.go
package dto

import (
	"encoding/base64"
	"time"

	"github.com/anoop-dryad/bridgehead/internal/downlink"
)

type CreateDownlinkRequest struct {
	ID         *string `json:"id,omitempty"`
	DeviceEUI  string  `json:"device_eui"  binding:"required"`
	DeviceType string  `json:"device_type" binding:"required,oneof=gateway sensor"`
	Payload    string  `json:"payload"     binding:"required"` // base64 encoded
	Type       string  `json:"type"        binding:"required,oneof=config command firmware ack"`
	ExpiresAt  *string `json:"expires_at,omitempty"`
}

type DownlinkResponse struct {
	ID         string `json:"id"`
	DeviceEUI  string `json:"device_eui"`
	DeviceType string `json:"device_type"`
	Payload    string `json:"payload"` // base64 encoded
	Type       string `json:"type"`
	Status     string `json:"status"`
	RetryCount int    `json:"retry_count"`
	CreatedAt  string `json:"created_at"`
	ExpiresAt  string `json:"expires_at"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func (r CreateDownlinkRequest) ToModel() (downlink.CreateRequest, error) {
	payload, err := base64.StdEncoding.DecodeString(r.Payload)
	if err != nil {
		return downlink.CreateRequest{}, err
	}

	req := downlink.CreateRequest{
		ID:         r.ID,
		DeviceEUI:  r.DeviceEUI,
		DeviceType: downlink.DeviceType(r.DeviceType),
		Payload:    payload,
		Type:       downlink.Type(r.Type),
	}

	if r.ExpiresAt != nil {
		t, err := time.Parse(time.RFC3339, *r.ExpiresAt)
		if err != nil {
			return downlink.CreateRequest{}, err
		}
		req.ExpiresAt = &t
	}

	return req, nil
}

func FromDownlink(d *downlink.DownlinkRequest) DownlinkResponse {
	return DownlinkResponse{
		ID:         d.ID,
		DeviceEUI:  d.DeviceEUI,
		DeviceType: string(d.DeviceType),
		Payload:    base64.StdEncoding.EncodeToString(d.Payload),
		Type:       string(d.Type),
		Status:     string(d.Status),
		RetryCount: d.RetryCount,
		CreatedAt:  d.CreatedAt.Format(time.RFC3339),
		ExpiresAt:  d.ExpiresAt.Format(time.RFC3339),
	}
}
