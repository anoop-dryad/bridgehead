package routing

import "context"

// routing declares only what it needs — not the whole sensor/gateway service
type SensorMappings interface {
	GetSensorsByGatewayEUI(ctx context.Context, bgEUI string) ([]MappedDevice, error)
	GetGatewayBySensorEUI(ctx context.Context, sensorEUI string) (string, error)
	GetSensorDetails(ctx context.Context, eui string) (SensorDetails, error)
}

type GatewayMappings interface {
	GetMeshGatewaysByBG(ctx context.Context, bgEUI string) ([]MappedDevice, error)
	GetBGByMeshEUI(ctx context.Context, mgEUI string) (string, error)
	GetKind(ctx context.Context, eui string) (Kind, error)
	GetGatewayID(ctx context.Context, eui string) (string, error) // TTN gateway_id
}

type SensorDetails struct {
	AppID    string
	DeviceID string
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

type DispatchTarget struct {
	Kind      Kind
	BGEUI     string // gateway to route through (border=itself, mesh/sensor=resolved)
	GatewayID string // BG's TTN gateway_id, for class_b_c
	AppID     string // sensor only
	DeviceID  string // sensor only
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

// ResolveDispatchTarget answers the one question the dispatcher asks:
// "everything I need to send this downlink to its target."
func (r *Resolver) ResolveDispatchTarget(ctx context.Context, deviceEUI string, kind Kind) (*DispatchTarget, error) {
	switch kind {
	case KindBorder:
		// the BG is the target itself
		gwID, err := r.gateways.GetGatewayID(ctx, deviceEUI)
		if err != nil {
			return nil, err
		}
		return &DispatchTarget{
			Kind:      KindBorder,
			BGEUI:     deviceEUI,
			GatewayID: gwID,
		}, nil

	case KindMesh:
		bgEUI, err := r.gateways.GetBGByMeshEUI(ctx, deviceEUI)
		if err != nil {
			return nil, err
		}
		gwID, err := r.gateways.GetGatewayID(ctx, bgEUI)
		if err != nil {
			return nil, err
		}
		return &DispatchTarget{
			Kind:      KindMesh,
			BGEUI:     bgEUI,
			GatewayID: gwID,
		}, nil

	case KindSensor:
		bgEUI, err := r.sensors.GetGatewayBySensorEUI(ctx, deviceEUI)
		if err != nil {
			return nil, err
		}
		details, err := r.sensors.GetSensorDetails(ctx, deviceEUI) // app_id, device_id
		if err != nil {
			return nil, err
		}
		gwID, err := r.gateways.GetGatewayID(ctx, bgEUI)
		if err != nil {
			return nil, err
		}
		return &DispatchTarget{
			Kind:      KindSensor,
			BGEUI:     bgEUI,
			GatewayID: gwID,
			AppID:     details.AppID,
			DeviceID:  details.DeviceID,
		}, nil

	default:
		return nil, ErrUnknownKind
	}
}
