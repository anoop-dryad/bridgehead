// internal/sensor/service.go
package sensor

import (
	"context"

	"go.uber.org/zap"
)

type Service struct {
	repo *Repository
	log  *zap.Logger
}

func NewService(repo *Repository, log *zap.Logger) *Service {
	return &Service{
		repo: repo,
		log:  log.With(zap.String("domain", "sensor")),
	}
}

func (s *Service) RecordUplink(ctx context.Context, event UplinkEvent) error {
	if event.GatewayEUI == "" {
		return ErrNoGatewayInUplink
	}

	return s.repo.WithTransaction(ctx, func(ctx context.Context) error {
		if err := s.repo.UpsertSensor(ctx, Sensor{
			EUI:      event.SensorEUI,
			DeviceID: event.DeviceID,
			AppID:    event.AppID,
		}); err != nil {
			s.log.Error("failed to upsert sensor",
				zap.String("eui", event.SensorEUI),
				zap.Error(err),
			)
			return err
		}

		if err := s.repo.UpsertMapping(ctx, GatewayMapping{
			SensorEUI:  event.SensorEUI,
			GatewayEUI: event.GatewayEUI,
		}); err != nil {
			s.log.Error("failed to upsert sensor gateway mapping",
				zap.String("sensor_eui", event.SensorEUI),
				zap.String("gateway_eui", event.GatewayEUI),
				zap.Error(err),
			)
			return err
		}

		s.log.Debug("sensor uplink recorded",
			zap.String("sensor_eui", event.SensorEUI),
			zap.String("gateway_eui", event.GatewayEUI),
		)
		return nil
	})
}

func (s *Service) GetByEUI(ctx context.Context, eui string) (*Sensor, error) {
	return s.repo.GetByEUI(ctx, eui)
}

func (s *Service) GetMappingBySensorEUI(ctx context.Context, sensorEUI string) (*GatewayMapping, error) {
	return s.repo.GetMappingBySensorEUI(ctx, sensorEUI)
}

func (s *Service) GetSensorsByGatewayEUI(ctx context.Context, gatewayEUI string) ([]*Sensor, error) {
	return s.repo.GetSensorsByGatewayEUI(ctx, gatewayEUI)
}
