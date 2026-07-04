package gateway

import (
	"fmt"
	"strings"
)

// topic pattern: /{bgeui}/u/{commandType}
type CommandType string

const (
	CommandTypeRPL        CommandType = "rpl"
	CommandTypeMQTTStatus CommandType = "mqttStatus"
)

type RPLPayload struct {
	Type string `json:"type"`
	Dest int64  `json:"dest"`
	Next int64  `json:"next"`
}

// mqttStatus — BG liveness
type MQTTStatusPayload struct {
	Status    string `json:"status"` // online/offline
	Timestamp int64  `json:"timestamp"`
}

// parsed from topic /{bgeui}/u/{commandType}
type TopicInfo struct {
	BGEUI       string
	CommandType CommandType
}

func ParseTopic(topic string) (*TopicInfo, error) {
	parts := strings.Split(topic, "/")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid topic format: %s", topic)
	}
	return &TopicInfo{
		BGEUI:       parts[0],
		CommandType: CommandType(parts[2]),
	}, nil
}
