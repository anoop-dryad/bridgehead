package dto

type CreateDownlinkRequest struct {
	DeviceID string `json:"device_id"`
	Payload  string `json:"payload"`
}
type DownlinkResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}
