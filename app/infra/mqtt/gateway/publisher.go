package gateway

import (
	"context"
	"fmt"
	"strings"

	appconfig "github.com/anoop-dryad/bridgehead/app/config"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"go.uber.org/zap"
)

type Publisher struct {
	client mqtt.Client
	log    *zap.Logger
}

func NewPublisher(cfg appconfig.GatewayMQTT, log *zap.Logger) (*Publisher, error) {
	opts := mqtt.NewClientOptions().
		AddBroker(cfg.BrokerURL).
		SetClientID(cfg.ClientID + "-pub").
		SetAutoReconnect(true)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("gateway publisher connect: %w", token.Error())
	}

	return &Publisher{
		client: client,
		log:    log.With(zap.String("infra", "mqtt-gw-pub")),
	}, nil
}

// Publish sends a downlink to a border/mesh gateway
// topic: /{eui}/d/{command}
func (p *Publisher) Publish(ctx context.Context, eui, command string, payload []byte) error {
	topic := fmt.Sprintf("/%s/d/%s", eui, strings.ToLower(command))

	token := p.client.Publish(topic, 0, false, payload)
	if token.Wait() && token.Error() != nil {
		return fmt.Errorf("gateway publish failed: %w", token.Error())
	}

	p.log.Debug("gateway downlink published",
		zap.String("eui", eui),
		zap.String("command", command))
	return nil
}

func (p *Publisher) Disconnect() {
	p.client.Disconnect(250)
}
