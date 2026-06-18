package cache

import (
	"time"

	model "github.com/anoop-dryad/bridgehead/internal/downlink"
	"github.com/hashicorp/golang-lru/v2/expirable"
)

// CacheService acts as the central hub for all in-memory data
type CacheService struct {
	UserCache *expirable.LRU[string, *model.User]
	/* Orgs    *expirable.LRU[string, *models.Org]
	Devices *expirable.LRU[string, *models.Device] */
}

func NewCacheService() *CacheService {
	return &CacheService{
		// 1-minute TTL for users (security kill-switch)
		UserCache: expirable.NewLRU[string, *model.User](10000, nil, time.Minute),

		/* Orgs: expirable.NewLRU[string, *models.Org](1000, nil, 10*time.Minute),
		Devices: expirable.NewLRU[string, *models.Device](5000, nil, 5*time.Minute), */
	}
}
