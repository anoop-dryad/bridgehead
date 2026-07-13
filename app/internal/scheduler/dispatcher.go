package scheduler

import (
	"context"

	"go.uber.org/zap"

	"github.com/anoop-dryad/bridgehead/app/internal/downlink"
	"github.com/anoop-dryad/bridgehead/app/internal/gateway"
	"github.com/anoop-dryad/bridgehead/app/internal/routing"
	"github.com/anoop-dryad/bridgehead/app/internal/sensor"
)

type GatewayPublisher interface {
	Publish(ctx context.Context, eui, command string, payload []byte) error
}

type TTNPublisher interface {
	Publish(ctx context.Context, appID, deviceID, frmPayload string, gatewayIDs []string) error
}

type Dispatcher struct {
	downlink   *downlink.Service
	sensor     *sensor.Service
	gateway    *gateway.Service
	gatewayPub GatewayPublisher
	ttnPub     TTNPublisher
	resolver   *routing.Resolver
	log        *zap.Logger
}

func NewDispatcher(
	dl *downlink.Service,
	sn *sensor.Service,
	gw *gateway.Service,
	gatewayPub GatewayPublisher,
	ttnPub TTNPublisher,
	resolver *routing.Resolver,
	log *zap.Logger,
) *Dispatcher {
	return &Dispatcher{
		downlink:   dl,
		sensor:     sn,
		gateway:    gw,
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
	// resolve LIVE — which targets currently route through this BG
	targetEUIs, err := d.resolver.ResolveTargets(ctx, bgEUI)
	if err != nil {
		d.log.Error("failed to resolve targets for bg",
			zap.String("bg_eui", bgEUI), zap.Error(err))
		return
	}
	if len(targetEUIs) == 0 {
		return
	}

	// claim queued downlinks — SKIP LOCKED prevents double-send across
	// concurrent flushes triggered by rapid successive uplinks
	requests, err := d.downlink.ClaimQueuedForTargets(ctx, targetEUIs)
	if err != nil {
		d.log.Error("failed to claim queued downlinks",
			zap.String("bg_eui", bgEUI), zap.Error(err))
		return
	}

	for _, req := range requests {
		d.dispatch(ctx, req, bgEUI)
	}
}

// dispatch — route one downlink to the correct publisher by target kind
func (d *Dispatcher) dispatch(ctx context.Context, req *downlink.DownlinkRequest, bgEUI string) {
	var err error

	switch req.DeviceType {
	case downlink.DeviceTypeGateway:
		// border or mesh — both go via gateway broker to the BG
		err = d.gatewayPub.Publish(ctx, bgEUI, string(req.Type), req.Payload)

	case downlink.DeviceTypeSensor:
		err = d.dispatchSensor(ctx, req, bgEUI)

	default:
		d.log.Error("unknown device type, marking failed",
			zap.String("id", req.ID),
			zap.String("device_type", string(req.DeviceType)))
		d.downlink.MarkFailed(ctx, req.ID)
		return
	}

	if err != nil {
		// publish failed — BG likely went silent. Leave QUEUED for next uplink.
		d.log.Error("dispatch failed, re-queueing",
			zap.String("id", req.ID),
			zap.String("bg_eui", bgEUI),
			zap.Error(err))
		d.downlink.Requeue(ctx, req.ID)
		return
	}

	d.downlink.MarkDispatched(ctx, req.ID)
	d.log.Debug("downlink dispatched",
		zap.String("id", req.ID),
		zap.String("bg_eui", bgEUI))
}

func (d *Dispatcher) dispatchSensor(ctx context.Context, req *downlink.DownlinkRequest, bgEUI string) error {
	// resolve sensor → app_id + device_id
	sn, err := d.sensor.GetByEUI(ctx, req.DeviceEUI)
	if err != nil {
		return err
	}

	// resolve BG → its TTN gateway_id for class_b_c routing
	bg, err := d.gateway.GetByEUI(ctx, bgEUI)
	if err != nil {
		return err
	}

	// payload is base64 in DB; TTN wants base64 string
	frmPayload := string(req.Payload)

	return d.ttnPub.Publish(ctx, sn.AppID, sn.DeviceID, frmPayload, []string{bg.GatewayID})
}
