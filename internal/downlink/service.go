// internal/downlink/service.go
package downlink

import "context"

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, req CreateRequest) (*DownlinkRequest, error) {
	if !isValidDeviceType(req.DeviceType) {
		return nil, ErrInvalidDeviceType
	}
	if !isValidType(req.Type) {
		return nil, ErrInvalidDownlinkType
	}
	return s.repo.Create(ctx, req)
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
	return s.repo.Delete(ctx, id)
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
