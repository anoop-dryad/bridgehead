// internal/downlink/service.go
package downlink

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
		log:  log.With(zap.String("domain", "downlink")), // scoped logger
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
	if req.Status == StatusDispatched || req.Status == StatusDelivered {
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
