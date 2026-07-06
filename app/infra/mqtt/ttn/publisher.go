package ttn

import (
	"context"
	"encoding/json"
	"fmt"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"go.uber.org/zap"

	appconfig "github.com/anoop-dryad/bridgehead/app/config"
)

const (
	sensorFPort     = 4
	defaultPriority = "NORMAL"
)

type Publisher struct {
	clients map[string]mqtt.Client // appID → persistent connection
	log     *zap.Logger
}

// TTN downlink message format
type downlinkMessage struct {
	Downlinks []downlink `json:"downlinks"`
}

type downlink struct {
	FrmPayload string  `json:"frm_payload"` // base64
	FPort      int     `json:"f_port"`
	Priority   string  `json:"priority"`
	ClassBC    classBC `json:"class_b_c"`
}

type classBC struct {
	Gateways []gatewayRef `json:"gateways"`
}

type gatewayRef struct {
	GatewayIDs gatewayID `json:"gateway_ids"`
}

type gatewayID struct {
	GatewayID string `json:"gateway_id"`
}

func NewPublisher(cfg appconfig.TTNMQTT, log *zap.Logger) (*Publisher, error) {
	clients := make(map[string]mqtt.Client)

	for _, app := range cfg.Apps {
		opts := mqtt.NewClientOptions().
			AddBroker(cfg.BrokerURL).
			SetClientID(fmt.Sprintf("bridgehead-ttn-%s", app.AppID)).
			SetUsername(app.Username).
			SetPassword(app.Password).
			SetAutoReconnect(true).
			SetConnectionLostHandler(func(_ mqtt.Client, err error) {
				log.Error("ttn connection lost",
					zap.String("app_id", app.AppID),
					zap.Error(err),
				)
			})

		client := mqtt.NewClient(opts)
		if token := client.Connect(); token.Wait() && token.Error() != nil {
			return nil, fmt.Errorf("ttn connect failed for app %s: %w", app.AppID, token.Error())
		}

		clients[app.AppID] = client
		log.Info("ttn connected", zap.String("app_id", app.AppID))
	}

	return &Publisher{
		clients: clients,
		log:     log.With(zap.String("infra", "mqtt-ttn")),
	}, nil
}

// Publish sends a downlink to a sensor via its TTN application
// topic: /v3/{appid}/{deviceid}/down/publish
func (p *Publisher) Publish(ctx context.Context, appID, deviceID string, frmPayload string, gatewayIDs []string) error {
	client, ok := p.clients[appID]
	if !ok {
		return fmt.Errorf("no ttn connection for app_id: %s", appID)
	}

	gateways := make([]gatewayRef, 0, len(gatewayIDs))
	for _, id := range gatewayIDs {
		gateways = append(gateways, gatewayRef{GatewayIDs: gatewayID{GatewayID: id}})
	}

	msg := downlinkMessage{
		Downlinks: []downlink{{
			FPort:      sensorFPort,
			FrmPayload: frmPayload,
			Priority:   defaultPriority,
			ClassBC:    classBC{Gateways: gateways},
		}},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal downlink: %w", err)
	}

	topic := fmt.Sprintf("/v3/%s/%s/down/replace", appID, deviceID)
	token := client.Publish(topic, 1, false, data)
	if token.Wait() && token.Error() != nil {
		return fmt.Errorf("ttn publish failed: %w", token.Error())
	}

	p.log.Debug("sensor downlink published",
		zap.String("app_id", appID),
		zap.String("device_id", deviceID),
		zap.Strings("gateways", gatewayIDs))
	return nil
}

func (p *Publisher) Disconnect() {
	for appID, client := range p.clients {
		client.Disconnect(250)
		p.log.Info("ttn disconnected", zap.String("app_id", appID))
	}
}
