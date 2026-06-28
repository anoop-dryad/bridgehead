package sensor

import "time"

type Sensor struct {
	ID        string
	EUI       string
	DeviceID  string
	AppID     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type GatewayMapping struct {
	SensorEUI  string
	GatewayEUI string
	UpdatedAt  time.Time
}

// UplinkEvent — what kinesis consumer passes to domain
type UplinkEvent struct {
	SensorEUI  string
	DeviceID   string
	AppID      string
	GatewayEUI string // already resolved to best gateway by infra layer
}
