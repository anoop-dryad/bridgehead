package sqs

type EventType string

const (
	EventDeviceRegistered EventType = "DeviceRegistered"
	EventDeviceDeleted    EventType = "DeviceDeleted"
)

type DeviceType string

const (
	DeviceTypeSensor  DeviceType = "sensor"
	DeviceTypeGateway DeviceType = "gateway"
)

type deviceEvent struct {
	EventType     EventType  `json:"event_type"`
	EUI           string     `json:"eui"`
	DeviceType    DeviceType `json:"device_type"`
	DeviceID      string     `json:"device_id"`       // sensor only
	AppID         string     `json:"app_id"`          // sensor only
	Kind          string     `json:"kind"`            // gateway only: bg/mg
	GatewayID     string     `json:"gateway_id"`      // gateway only
	SiteGatewayID int64      `json:"site_gateway_id"` // gateway only
}
