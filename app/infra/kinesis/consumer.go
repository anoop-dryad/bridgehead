// infra/kinesis/consumer.go
package kinesis

import (
	"context"
	"encoding/json"
	"strings"

	appconfig "github.com/anoop-dryad/bridgehead/app/config"
	"github.com/anoop-dryad/bridgehead/app/internal/sensor"
	consumer "github.com/harlow/kinesis-consumer"
	store "github.com/harlow/kinesis-consumer/store/postgres"
	"go.uber.org/zap"
)

type Consumer struct {
	consumer      *consumer.Consumer // one stream per consumer
	sensorService *sensor.Service
	log           *zap.Logger
}

func NewConsumer(cfg appconfig.Kinesis, svc *sensor.Service, log *zap.Logger) (*Consumer, error) {
	ck, err := store.New("bridgehead", "kinesis_consumer", cfg.DSN)
	if err != nil {
		return nil, err
	}

	active_consumer, err := consumer.New(
		cfg.StreamName,
		consumer.WithStore(ck),
		consumer.WithLogger(&zapAdapter{log: log}),
	)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		consumer:      active_consumer,
		sensorService: svc,
		log:           log.With(zap.String("infra", "kinesis")),
	}, nil
}

func (c *Consumer) Start(ctx context.Context) error {
	c.log.Info("starting kinesis consumer")

	return c.consumer.Scan(ctx, func(r *consumer.Record) error {
		c.handleRecord(ctx, r.Data)
		return nil // continue scanning
	})
}

func (c *Consumer) handleRecord(ctx context.Context, data []byte) {
	var uplink TTIUplink
	if err := json.Unmarshal(data, &uplink); err != nil {
		c.log.Error("failed to decode kinesis record", zap.Error(err))
		return
	}

	best := bestGateway(uplink.UplinkMessage.RxMetadata)
	if best == nil {
		c.log.Error("no gateway metadata in uplink",
			zap.String("dev_eui", uplink.EndDeviceIDs.DevEUI),
		)
		return
	}

	if err := c.sensorService.RecordUplink(ctx, sensor.UplinkEvent{
		SensorEUI:  strings.ToLower(uplink.EndDeviceIDs.DevEUI),
		DeviceID:   strings.ToLower(uplink.EndDeviceIDs.DeviceID),
		AppID:      uplink.EndDeviceIDs.AppIDs.ApplicationID,
		GatewayEUI: strings.ToLower(best.GatewayIDs.EUI),
	}); err != nil {
		c.log.Error("failed to record uplink",
			zap.String("dev_eui", uplink.EndDeviceIDs.DevEUI),
			zap.Error(err),
		)
	}
}

// zapAdapter bridges harlow's Logger interface to zap
type zapAdapter struct {
	log *zap.Logger
}

func (z *zapAdapter) Log(args ...interface{}) {
	z.log.Sugar().Info(args...)
}
