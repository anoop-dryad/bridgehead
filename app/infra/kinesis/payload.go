package kinesis

import "time"

type TTIUplink struct {
	ReceivedAt    time.Time     `json:"received_at"`
	EndDeviceIDs  EndDeviceIDs  `json:"end_device_ids"`
	UplinkMessage UplinkMessage `json:"uplink_message"`
}

type EndDeviceIDs struct {
	DevEUI   string         `json:"dev_eui"`
	DeviceID string         `json:"device_id"`
	AppIDs   ApplicationIDs `json:"application_ids"`
}

type ApplicationIDs struct {
	ApplicationID string `json:"application_id"`
}

type UplinkMessage struct {
	RxMetadata []RxMetadata `json:"rx_metadata"`
}

type RxMetadata struct {
	GatewayIDs GatewayIDs `json:"gateway_ids"`
	RSSI       float64    `json:"rssi"`
	ReceivedAt time.Time  `json:"received_at"`
}

type GatewayIDs struct {
	EUI       string `json:"eui"`
	GatewayID string `json:"gateway_id"`
}

// bestGateway picks the gateway with highest RSSI
// RSSI is negative — closer to zero = stronger signal
func bestGateway(metadata []RxMetadata) *RxMetadata {
	if len(metadata) == 0 {
		return nil
	}
	best := &metadata[0]
	for i := range metadata {
		if metadata[i].RSSI > best.RSSI {
			best = &metadata[i]
		}
	}
	return best
}
