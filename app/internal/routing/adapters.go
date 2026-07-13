package routing

import (
	"context"

	"github.com/anoop-dryad/bridgehead/app/internal/gateway"
	"github.com/anoop-dryad/bridgehead/app/internal/sensor"
)

// sensorAdapter wraps sensor.Service to satisfy routing.SensorMappings
type sensorAdapter struct {
	svc *sensor.Service
}

type gatewayAdapter struct {
	svc *gateway.Service
}

func NewSensorAdapter(svc *sensor.Service) SensorMappings {
	return &sensorAdapter{svc: svc}
}

func NewGatewayAdapter(svc *gateway.Service) GatewayMappings {
	return &gatewayAdapter{svc: svc}
}

func (a *sensorAdapter) GetSensorsByGatewayEUI(ctx context.Context, bgEUI string) ([]MappedDevice, error) {
	sensors, err := a.svc.GetSensorsByGatewayEUI(ctx, bgEUI)
	if err != nil {
		return nil, err
	}
	out := make([]MappedDevice, len(sensors))
	for i, s := range sensors {
		out[i] = MappedDevice{EUI: s.EUI, Kind: KindSensor}
	}
	return out, nil
}

func (a *sensorAdapter) GetGatewayBySensorEUI(ctx context.Context, sensorEUI string) (string, error) {
	mapping, err := a.svc.GetMappingBySensorEUI(ctx, sensorEUI)
	if err != nil {
		return "", err
	}
	return mapping.GatewayEUI, nil
}

func (a *gatewayAdapter) GetMeshGatewaysByBG(ctx context.Context, bgEUI string) ([]MappedDevice, error) {
	mgs, err := a.svc.GetMeshGatewaysByBG(ctx, bgEUI)
	if err != nil {
		return nil, err
	}
	out := make([]MappedDevice, len(mgs))
	for i, mg := range mgs {
		out[i] = MappedDevice{EUI: mg.EUI, Kind: KindMesh}
	}
	return out, nil
}

func (a *gatewayAdapter) GetKind(ctx context.Context, eui string) (Kind, error) {
	gw, err := a.svc.GetByEUI(ctx, eui)
	if err != nil {
		return "", err
	}
	// translate gateway domain kind → routing kind
	switch gw.Kind {
	case gateway.TypeBG:
		return KindBorder, nil
	case gateway.TypeMG:
		return KindMesh, nil
	default:
		return "", ErrUnknownKind
	}
}

func (a *gatewayAdapter) GetBGByMeshEUI(ctx context.Context, mgEUI string) (string, error) {
	// ← this needs a service method that doesn't exist yet
	return a.svc.GetBGByMeshEUI(ctx, mgEUI)
}
