package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"

	appconfig "github.com/anoop-dryad/bridgehead/app/config"
	"github.com/anoop-dryad/bridgehead/app/internal/gateway"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Consumer struct {
	client         mqtt.Client
	gatewayService *gateway.Service
	log            *zap.Logger
}

func NewConsumer(cfg appconfig.GatewayMQTT, svc *gateway.Service, log *zap.Logger) (*Consumer, error) {
	opts := mqtt.NewClientOptions().
		AddBroker(cfg.BrokerURL).
		SetClientID(cfg.ClientID).
		SetAutoReconnect(true).
		SetConnectionLostHandler(func(c mqtt.Client, err error) {
			log.Error("gateway mqtt connection lost", zap.Error(err))
		}).
		SetOnConnectHandler(func(c mqtt.Client) {
			log.Info("gateway mqtt connected")
		})

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("gateway mqtt connect failed: %w", token.Error())
	}

	return &Consumer{
		client:         client,
		gatewayService: svc,
		log:            log.With(zap.String("infra", "mqtt-gateway")),
	}, nil
}

func (c *Consumer) Start(ctx context.Context) error {
	// wildcard — all gateways, all command types
	topic := "+/u/#"

	token := c.client.Subscribe(topic, 1, func(_ mqtt.Client, msg mqtt.Message) {
		c.handleMessage(msg)
	})
	if token.Wait() && token.Error() != nil {
		return fmt.Errorf("subscribe failed: %w", token.Error())
	}

	c.log.Info("subscribed to gateway uplinks", zap.String("topic", topic))

	<-ctx.Done()
	c.client.Disconnect(250)
	c.log.Info("gateway mqtt consumer stopped")
	return nil
}

func (c *Consumer) handleMessage(msg mqtt.Message) {
	info, err := ParseTopic(msg.Topic())
	if err != nil {
		c.log.Error("invalid topic", zap.String("topic", msg.Topic()), zap.Error(err))
		return
	}

	switch info.CommandType {
	case CommandTypeMQTTStatus:
		c.handleMQTTStatus(info.BGEUI, msg.Payload())
	case CommandTypeRPL:
		c.handleRPL(info.BGEUI, msg.Payload())
	default:
		c.log.Debug("unhandled command type",
			zap.String("type", string(info.CommandType)),
		)
	}
}

func (c *Consumer) handleMQTTStatus(bgeui string, data []byte) {
	var payload MQTTStatusPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		c.log.Error("failed to parse mqttStatus",
			zap.String("bgeui", bgeui),
			zap.Error(err),
		)
		return
	}

	// hand off to domain
	if err := c.gatewayService.RecordUplink(
		context.Background(), bgeui, time.Now(),
	); err != nil {
		c.log.Error("failed to record gateway uplink",
			zap.String("bgeui", bgeui),
			zap.Error(err),
		)
	}
}

func (c *Consumer) handleRPL(bgeui string, data []byte) {
	var payload RPLPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		c.log.Error("failed to parse rpl", zap.String("bgeui", bgeui), zap.Error(err))
		return
	}

	// only process DAO_PATH — ignore other rpl types
	if payload.Type != "DAO_PATH" {
		c.log.Debug("ignoring non-DAO_PATH rpl", zap.String("type", payload.Type))
		return
	}

	if err := c.gatewayService.RecordMeshRegistration(
		context.Background(), bgeui, payload.Dest,
	); err != nil {
		c.log.Error("failed to record mesh registration",
			zap.String("bgeui", bgeui),
			zap.Int64("dest", payload.Dest),
			zap.Error(err),
		)
	}
}
