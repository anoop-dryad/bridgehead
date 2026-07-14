package downlink

import "time"

type Type string
type DeviceType string

const (
	TypeConfig   Type = "config"
	TypeCommand  Type = "command"
	TypeFirmware Type = "firmware"
	TypeAck      Type = "ack"
)

const (
	DeviceTypeGateway DeviceType = "gateway"
	DeviceTypeSensor  DeviceType = "sensor"
)

const DefaultTTL = 24 * time.Hour

type DownlinkRequest struct {
	ID         string
	DeviceEUI  string
	DeviceType DeviceType
	Payload    []byte
	Type       Type
	Status     Status
	RetryCount int
	CreatedAt  time.Time
	UpdatedAt  time.Time
	ExpiresAt  time.Time
}

type CreateRequest struct {
	ID         *string // optional — caller may provide
	DeviceEUI  string
	DeviceType DeviceType
	Payload    []byte
	Type       Type
	ExpiresAt  *time.Time // optional — defaults to now + 24h
}
