// infra/kinesis/consumer.go
package kinesis

import (
	"context"
	"encoding/json"
	"time"

	appconfig "github.com/anoop-dryad/bridgehead/app/config"
	"github.com/anoop-dryad/bridgehead/app/internal/sensor"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	"go.uber.org/zap"
)

type Consumer struct {
	client        *kinesis.Client
	streamName    string
	sensorService *sensor.Service
	log           *zap.Logger
}

func NewConsumer(cfg appconfig.Kinesis, svc *sensor.Service, log *zap.Logger) (*Consumer, error) {
	awsCfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(cfg.Region),
	)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		client:        kinesis.NewFromConfig(awsCfg),
		streamName:    cfg.StreamName,
		sensorService: svc,
		log:           log.With(zap.String("infra", "kinesis")),
	}, nil
}

func (c *Consumer) Start(ctx context.Context) {
	c.log.Info("starting kinesis consumer",
		zap.String("stream", c.streamName),
	)

	for {
		if err := c.consume(ctx); err != nil {
			c.log.Error("kinesis consumer error, retrying in 5s",
				zap.Error(err),
			)
			select {
			case <-ctx.Done():
				return
			case <-time.After(5 * time.Second):
				continue
			}
		}
		if ctx.Err() != nil {
			return
		}
	}
}

func (c *Consumer) consume(ctx context.Context) error {
	// get shards
	streams, err := c.client.DescribeStream(ctx, &kinesis.DescribeStreamInput{
		StreamName: aws.String(c.streamName),
	})
	if err != nil {
		return err
	}

	for _, shard := range streams.StreamDescription.Shards {
		go c.consumeShard(ctx, aws.ToString(shard.ShardId))
	}

	<-ctx.Done()
	return nil
}

func (c *Consumer) consumeShard(ctx context.Context, shardID string) {
	// get shard iterator — LATEST means only new records
	iter, err := c.client.GetShardIterator(ctx, &kinesis.GetShardIteratorInput{
		StreamName:        aws.String(c.streamName),
		ShardId:           aws.String(shardID),
		ShardIteratorType: "LATEST",
	})
	if err != nil {
		c.log.Error("failed to get shard iterator",
			zap.String("shard", shardID),
			zap.Error(err),
		)
		return
	}

	nextIterator := iter.ShardIterator

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		out, err := c.client.GetRecords(ctx, &kinesis.GetRecordsInput{
			ShardIterator: nextIterator,
			Limit:         aws.Int32(100),
		})
		if err != nil {
			c.log.Error("failed to get records",
				zap.String("shard", shardID),
				zap.Error(err),
			)
			time.Sleep(1 * time.Second)
			continue
		}

		for _, record := range out.Records {
			c.handleRecord(ctx, record.Data)
		}

		nextIterator = out.NextShardIterator
		if nextIterator == nil {
			return // shard ended
		}

		// avoid hot loop when no records
		if len(out.Records) == 0 {
			time.Sleep(1 * time.Second)
		}
	}
}

func (c *Consumer) handleRecord(ctx context.Context, data []byte) {
	var uplink TTIUplink
	if err := json.Unmarshal(data, &uplink); err != nil {
		c.log.Error("failed to decode kinesis record", zap.Error(err))
		return
	}

	// pick best gateway — highest RSSI
	best := bestGateway(uplink.UplinkMessage.RxMetadata)
	if best == nil {
		c.log.Error("no gateway metadata in uplink",
			zap.String("dev_eui", uplink.EndDeviceIDs.DevEUI),
		)
		return
	}

	// hand off to domain — infra job is done
	err := c.sensorService.RecordUplink(ctx, sensor.UplinkEvent{
		SensorEUI:  uplink.EndDeviceIDs.DevEUI,
		DeviceID:   uplink.EndDeviceIDs.DeviceID,
		AppID:      uplink.EndDeviceIDs.AppIDs.ApplicationID,
		GatewayEUI: best.GatewayIDs.EUI,
	})
	if err != nil {
		c.log.Error("failed to record uplink",
			zap.String("dev_eui", uplink.EndDeviceIDs.DevEUI),
			zap.Error(err),
		)
	}
}
