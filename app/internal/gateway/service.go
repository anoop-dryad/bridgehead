package gateway

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

type Cache interface {
	Set(ctx context.Context, key string, val string, ttl time.Duration) error
	Exists(ctx context.Context, key string) (bool, error)
}

type Service struct {
	repo  RepositoryInterface
	cache Cache
	log   *zap.Logger
}

func NewService(repo RepositoryInterface, cache Cache, log *zap.Logger) *Service {
	return &Service{
		repo:  repo,
		cache: cache,
		log:   log.With(zap.String("domain", "gateway")),
	}
}

const (
	livenessTTL    = 30 * time.Second
	livenessPrefix = "gw:liveness:"
)

func livenessKey(eui string) string {
	return livenessPrefix + eui
}

// RecordUplink — called by mqtt consumer on mqttStatus uplink
// updates liveness in Redis only — no DB write
func (s *Service) RecordUplink(ctx context.Context, bgeui string, at time.Time) error {
	if err := s.cache.Set(ctx, livenessKey(bgeui), "online", livenessTTL); err != nil {
		s.log.Error("failed to update gateway liveness",
			zap.String("bgeui", bgeui),
			zap.Error(err),
		)
		return err
	}
	s.log.Debug("gateway liveness updated", zap.String("bgeui", bgeui))
	return nil
}

// IsOnline — called by dispatcher before sending downlink
func (s *Service) IsOnline(ctx context.Context, bgeui string) bool {
	exists, err := s.cache.Exists(ctx, livenessKey(bgeui))
	if err != nil {
		s.log.Error("failed to check gateway liveness",
			zap.String("bgeui", bgeui),
			zap.Error(err),
		)
		return false
	}
	return exists
}

// RecordMeshRegistration — called by mqtt consumer on rpl DAO_PATH uplink
// resolves numeric site_gateway_id → EUI then upserts mapping
func (s *Service) RecordMeshRegistration(ctx context.Context, bgEUI string, mgSiteGatewayID int64) error {
	mg, err := s.repo.GetBySiteGatewayID(ctx, mgSiteGatewayID)
	if err != nil {
		s.log.Error("mesh gateway not found",
			zap.String("bg_eui", bgEUI),
			zap.Int64("site_gateway_id", mgSiteGatewayID),
			zap.Error(err),
		)
		return fmt.Errorf("mesh gateway not found: %w", err)
	}

	if err := s.repo.UpsertMeshMapping(ctx, MeshMapping{
		BGEUI: bgEUI,
		MGEUI: mg.EUI,
	}); err != nil {
		s.log.Error("failed to upsert mesh mapping",
			zap.String("bg_eui", bgEUI),
			zap.String("mg_eui", mg.EUI),
			zap.Error(err),
		)
		return err
	}

	s.log.Debug("mesh registration recorded",
		zap.String("bg_eui", bgEUI),
		zap.String("mg_eui", mg.EUI),
	)
	return nil
}
