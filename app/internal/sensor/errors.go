package sensor

import "errors"

var (
	ErrNotFound          = errors.New("sensor not found")
	ErrMappingNotFound   = errors.New("sensor gateway mapping not found")
	ErrNoGatewayInUplink = errors.New("no gateway metadata in uplink")
)
