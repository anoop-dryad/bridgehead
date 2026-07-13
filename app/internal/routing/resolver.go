package routing

import "context"

// routing declares only what it needs — not the whole sensor/gateway service
type SensorMappings interface {
	GetSensorsByGatewayEUI(ctx context.Context, bgEUI string) ([]MappedDevice, error)
	GetGatewayBySensorEUI(ctx context.Context, sensorEUI string) (string, error)
}

type GatewayMappings interface {
	GetMeshGatewaysByBG(ctx context.Context, bgEUI string) ([]MappedDevice, error)
	GetBGByMeshEUI(ctx context.Context, mgEUI string) (string, error)
	GetKind(ctx context.Context, eui string) (Kind, error)
}

type MappedDevice struct {
	EUI  string
	Kind Kind
}

type Kind string

const (
	KindBorder Kind = "border"
	KindMesh   Kind = "mesh"
	KindSensor Kind = "sensor"
)

type Resolver struct {
	sensors  SensorMappings
	gateways GatewayMappings
}

func New(s SensorMappings, g GatewayMappings) *Resolver {
	return &Resolver{sensors: s, gateways: g}
}

// ResolveBG — device → the BG that reaches it
func (r *Resolver) ResolveBG(ctx context.Context, targetEUI string, targetKind Kind) (string, error) {
	switch targetKind {
	case KindBorder:
		return targetEUI, nil // the BG is the target itself
	case KindMesh:
		return r.gateways.GetBGByMeshEUI(ctx, targetEUI)
	case KindSensor:
		return r.sensors.GetGatewayBySensorEUI(ctx, targetEUI)
	default:
		return "", ErrUnknownKind
	}
}

// ResolveTargets — BG → all devices currently routing through it
func (r *Resolver) ResolveTargets(ctx context.Context, bgEUI string) ([]string, error) {
	targets := []string{bgEUI} // border: the BG itself

	sensors, err := r.sensors.GetSensorsByGatewayEUI(ctx, bgEUI)
	if err != nil {
		return nil, err
	}
	for _, s := range sensors {
		targets = append(targets, s.EUI)
	}

	mgs, err := r.gateways.GetMeshGatewaysByBG(ctx, bgEUI)
	if err != nil {
		return nil, err
	}
	for _, mg := range mgs {
		targets = append(targets, mg.EUI)
	}

	return targets, nil
}
