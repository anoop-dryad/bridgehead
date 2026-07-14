// internal/downlink/service.go
package downlink

import (
	"context"
	"time"

	"github.com/anoop-dryad/bridgehead/app/internal/routing"
	"go.uber.org/zap"
)

type GatewayProbe interface {
	Publish(ctx context.Context, eui, command string, payload []byte) error
}

type Service struct {
	repo     RepositoryInterface
	probe    GatewayProbe
	resolver *routing.Resolver
	log      *zap.Logger
}

const probeCommand = "liveness"

func NewService(repo RepositoryInterface, resolver *routing.Resolver, log *zap.Logger) *Service {
	return &Service{
		repo:     repo,
		resolver: resolver,
		log:      log.With(zap.String("domain", "downlink")), // scoped logger
	}
}

func (s *Service) Create(ctx context.Context, req CreateRequest) (*DownlinkRequest, error) {
	if !isValidDeviceType(req.DeviceType) {
		return nil, ErrInvalidDeviceType
	}
	if !isValidType(req.Type) {
		return nil, ErrInvalidDownlinkType
	}

	result, err := s.repo.Create(ctx, req)
	if err != nil {
		s.log.Error("failed to create downlink request",
			zap.String("device_eui", req.DeviceEUI),
			zap.Error(err),
		)
		return nil, err
	}

	// resolve target's BG and probe it once — poke it awake.
	// online  → BG emits uplink → dispatcher.FlushBG fires
	// offline → probe lost, harmless
	bgEUI, err := s.resolver.ResolveBG(ctx, result.DeviceEUI, routing.Kind(req.DeviceType))
	if err == nil && bgEUI != "" {
		if perr := s.probe.Publish(ctx, bgEUI, probeCommand, nil); perr != nil {
			s.log.Debug("probe failed (harmless)",
				zap.String("bg_eui", bgEUI), zap.Error(perr))
		}
	}

	return result, nil
}

func (s *Service) GetByID(ctx context.Context, id string) (*DownlinkRequest, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	req, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	// only allow delete if not already dispatched
	if req.Status != StatusDispatched {
		return ErrInvalidStatus
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		s.log.Error("failed to delete downlink request",
			zap.String("id", id),
			zap.Error(err),
		)
		return err
	}
	return nil
}

func (s *Service) List(ctx context.Context, deviceEUI string) ([]*DownlinkRequest, error) {
	return s.repo.List(ctx, deviceEUI)
}

func (s *Service) ClaimQueuedForTargets(ctx context.Context, targetEUIs []string) ([]*DownlinkRequest, error) {
	return s.repo.ClaimQueuedForTargets(ctx, targetEUIs)
}

func (s *Service) ExpireStale(ctx context.Context) (int64, error) {
	return s.repo.ExpireStale(ctx)
}

// MarkDispatched — publish succeeded.
func (s *Service) MarkDispatched(ctx context.Context, id string) error {
	return s.apply(ctx, id, DispatchSuccess)
}

// HandleFailure — publish failed; machine decides retry-with-backoff vs FAILED.
func (s *Service) HandleFailure(ctx context.Context, id string) error {
	return s.apply(ctx, id, DispatchFailed)
}

func (s *Service) MarkFailedPermanent(ctx context.Context, id string) error {
	return s.repo.UpdateStatus(ctx, id, StatusFailed)
}

func (s *Service) apply(ctx context.Context, id string, result DispatchResult) error {
	req, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// terminal states are immutable — guard against double-processing
	if IsTerminal(req.Status) {
		s.log.Debug("ignoring transition on terminal request",
			zap.String("id", id), zap.String("status", string(req.Status)))
		return nil
	}

	d := Decide(req, result, time.Now())

	if !CanTransition(req.Status, d.Next) {
		s.log.Error("illegal transition blocked",
			zap.String("id", id),
			zap.String("from", string(req.Status)),
			zap.String("to", string(d.Next)))
		return ErrInvalidStatus
	}

	if d.IncrementRetry {
		return s.repo.RequeueWithBackoff(ctx, id, d.NextEligibleAt)
	}
	return s.repo.UpdateStatus(ctx, id, d.Next)
}

func isValidDeviceType(d DeviceType) bool {
	return d == DeviceTypeGateway || d == DeviceTypeSensor
}

func isValidType(t Type) bool {
	switch t {
	case TypeConfig, TypeCommand, TypeFirmware, TypeAck:
		return true
	}
	return false
}
