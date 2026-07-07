package sqs

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	appconfig "github.com/anoop-dryad/bridgehead/app/config"
	"github.com/anoop-dryad/bridgehead/app/internal/gateway"
	"github.com/anoop-dryad/bridgehead/app/internal/sensor"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"go.uber.org/zap"
)

type Consumer struct {
	log            *zap.Logger
	client         *sqs.Client
	queueURL       string
	sensorService  *sensor.Service
	gatewayService *gateway.Service
}

func NewConsumer(cfg appconfig.SQS,
	sensorSvc *sensor.Service,
	gatewaySvc *gateway.Service,
	log *zap.Logger) (*Consumer, error) {

	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion(cfg.Region),
	)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		client:         sqs.NewFromConfig(awsCfg),
		queueURL:       cfg.QueueURL,
		sensorService:  sensorSvc,
		gatewayService: gatewaySvc,
		log:            log.With(zap.String("infra", "sqs")),
	}, nil

}

func (c *Consumer) Start(ctx context.Context) error {
	c.log.Info("starting sqs consumer", zap.String("queue", c.queueURL))

	for {

		select {
		case <-ctx.Done():
			c.log.Info("sqs consumer stopped")
			return nil
		default:
		}

		out, err := c.client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
			QueueUrl:            aws.String(c.queueURL),
			MaxNumberOfMessages: 10,
			WaitTimeSeconds:     20, // long polling
		})
		if err != nil {
			c.log.Error("failed to receive messages", zap.Error(err))
			time.Sleep(5 * time.Second)
			continue
		}

		for _, msg := range out.Messages {
			if c.handleMessage(ctx, *msg.Body) {
				// delete only on successful processing
				_, err := c.client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
					QueueUrl:      aws.String(c.queueURL),
					ReceiptHandle: msg.ReceiptHandle,
				})
				if err != nil {
					c.log.Error("failed to delete message", zap.Error(err))
				}
			}
		}
	}
}

// returns true if processed successfully (safe to delete from queue)
func (c *Consumer) handleMessage(ctx context.Context, body string) bool {
	var event deviceEvent
	if err := json.Unmarshal([]byte(body), &event); err != nil {
		c.log.Error("failed to decode sqs message", zap.Error(err))
		return true // malformed — delete, don't retry forever
	}

	switch event.DeviceType {
	case DeviceTypeSensor:
		return c.handleSensor(ctx, event)
	case DeviceTypeGateway:
		return c.handleGateway(ctx, event)
	default:
		c.log.Error("unknown device type", zap.String("type", string(event.DeviceType)))
		return true // unknown — delete
	}
}

func (c *Consumer) handleSensor(ctx context.Context, e deviceEvent) bool {
	if e.EventType == EventDeviceDeleted {
		if err := c.sensorService.Delete(ctx, strings.ToLower(e.EUI)); err != nil {
			c.log.Error("failed to delete sensor", zap.String("eui", e.EUI), zap.Error(err))
			return false // retry
		}
		return true
	}

	// registered — upsert
	if err := c.sensorService.Upsert(ctx, sensor.Sensor{
		EUI:      strings.ToLower(e.EUI),
		DeviceID: strings.ToLower(e.DeviceID),
		AppID:    e.AppID,
	}); err != nil {
		c.log.Error("failed to upsert sensor", zap.String("eui", e.EUI), zap.Error(err))
		return false // retry
	}
	return true
}

func (c *Consumer) handleGateway(ctx context.Context, e deviceEvent) bool {
	if e.EventType == EventDeviceDeleted {
		if err := c.gatewayService.Delete(ctx, strings.ToLower(e.EUI)); err != nil {
			c.log.Error("failed to delete gateway", zap.String("eui", e.EUI), zap.Error(err))
			return false
		}
		return true
	}

	if err := c.gatewayService.Upsert(ctx, gateway.Gateway{
		EUI:           strings.ToLower(e.EUI),
		GatewayID:     strings.ToLower(e.GatewayID),
		SiteGatewayID: e.SiteGatewayID,
		Kind:          gateway.Kind(e.Kind),
	}); err != nil {
		c.log.Error("failed to upsert gateway", zap.String("eui", e.EUI), zap.Error(err))
		return false
	}
	return true
}
