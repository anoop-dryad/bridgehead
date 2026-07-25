package scheduler

import (
	"context"

	"go.uber.org/zap"

	"github.com/anoop-dryad/bridgehead/app/internal/downlink"
	"github.com/anoop-dryad/bridgehead/app/internal/routing"
)

type GatewayPublisher interface {
	Publish(ctx context.Context, eui, command string, payload []byte) error
}

type TTNPublisher interface {
	Publish(ctx context.Context, appID, deviceID, frmPayload string, gatewayIDs []string) error
}

type Dispatcher struct {
	downlink   *downlink.Service
	gatewayPub GatewayPublisher
	ttnPub     TTNPublisher
	resolver   *routing.Resolver
	log        *zap.Logger
}

func NewDispatcher(
	dl *downlink.Service,
	gatewayPub GatewayPublisher,
	ttnPub TTNPublisher,
	resolver *routing.Resolver,
	log *zap.Logger,
) *Dispatcher {
	return &Dispatcher{
		downlink:   dl,
		gatewayPub: gatewayPub,
		ttnPub:     ttnPub,
		resolver:   resolver,
		log:        log.With(zap.String("component", "dispatcher")),
	}
}

// FlushBG — called by gateway consumer on ANY uplink from bgEUI.
// Finds all QUEUED downlinks currently routed through this BG and sends them.
// Concurrent calls are safe: FOR UPDATE SKIP LOCKED ensures each downlink
// is claimed by exactly one flush.
func (d *Dispatcher) FlushBG(ctx context.Context, bgEUI string) {
	targetEUIs, err := d.resolver.ResolveTargets(ctx, bgEUI)
	if err != nil {
		d.log.Error("failed to resolve targets for bg",
			zap.String("bg_eui", bgEUI), zap.Error(err))
		return
	}
	if len(targetEUIs) == 0 {
		return
	}

	requests, err := d.downlink.ClaimQueuedForTargets(ctx, targetEUIs)
	if err != nil {
		d.log.Error("failed to claim queued downlinks",
			zap.String("bg_eui", bgEUI), zap.Error(err))
		return
	}

	for _, req := range requests {
		d.dispatch(ctx, req)
	}
}

// dispatch — resolve full addressing via the resolver, then publish.
func (d *Dispatcher) dispatch(ctx context.Context, req *downlink.DownlinkRequest) {
	kind, ok := mapDeviceKind(req.DeviceType)
	if !ok {
		d.log.Error("unknown device type, marking failed permanently",
			zap.String("id", req.ID),
			zap.String("device_type", string(req.DeviceType)))
		d.downlink.MarkFailedPermanent(ctx, req.ID)
		return
	}

	target, err := d.resolver.ResolveDispatchTarget(ctx, req.DeviceEUI, kind)
	if err != nil {
		// can't resolve routing right now — leave QUEUED, retry on next uplink
		d.log.Error("failed to resolve dispatch target, re-queueing",
			zap.String("id", req.ID),
			zap.String("device_eui", req.DeviceEUI),
			zap.Error(err))
		d.downlink.HandleFailure(ctx, req.ID)
		return
	}

	switch target.Kind {
	case routing.KindBorder, routing.KindMesh:
		err = d.gatewayPub.Publish(ctx, target.BGEUI, string(req.Type), req.Payload)
	case routing.KindSensor:
		// payload is base64 in DB; TTN wants the base64 string as-is
		err = d.ttnPub.Publish(ctx, target.AppID, target.DeviceID, string(req.Payload), []string{target.GatewayID})
	}

	if err != nil {
		d.log.Error("dispatch failed, re-queueing",
			zap.String("id", req.ID),
			zap.String("bg_eui", target.BGEUI),
			zap.Error(err))
		d.downlink.HandleFailure(ctx, req.ID)
		return
	}

	d.downlink.MarkDispatched(ctx, req.ID)
	d.log.Debug("downlink dispatched",
		zap.String("id", req.ID),
		zap.String("bg_eui", target.BGEUI))
}

// mapDeviceKind maps the coarse downlink DeviceType to a routing.Kind.
//
// TODO: DeviceType (sensor|gateway) is coarser than routing.Kind
// (border|mesh|sensor). A gateway downlink is treated as KindBorder here,
// which is WRONG for mesh targets. Revisit: either store border/mesh on the
// request at creation, or have the resolver refine gateway→border/mesh via
// GetKind. See dispatcher discussion.
func mapDeviceKind(dt downlink.DeviceType) (routing.Kind, bool) {
	switch dt {
	case downlink.DeviceTypeSensor:
		return routing.KindSensor, true
	case downlink.DeviceTypeGateway:
		return routing.KindBorder, true // ← temporary: mesh mis-routed as border
	default:
		return "", false
	}
}
