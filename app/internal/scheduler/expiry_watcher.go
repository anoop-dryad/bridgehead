package scheduler

import (
	"context"
	"time"

	"github.com/anoop-dryad/bridgehead/app/internal/downlink"
	"go.uber.org/zap"
)

type ExpiryWatcher struct {
	downlink *downlink.Service
	log      *zap.Logger
}

func NewExpiryWatcher(dl *downlink.Service, log *zap.Logger) *ExpiryWatcher {
	return &ExpiryWatcher{
		downlink: dl,
		log:      log.With(zap.String("component", "expiry-watcher")),
	}
}

func (w *ExpiryWatcher) Run(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	// sweep once at startup — catch anything stale from downtime
	w.sweep(ctx)

	for {
		select {
		case <-ctx.Done():
			w.log.Info("expiry watcher stopped")
			return
		case <-ticker.C:
			w.sweep(ctx)
		}
	}
}

func (w *ExpiryWatcher) sweep(ctx context.Context) {
	n, err := w.downlink.ExpireStale(ctx)
	if err != nil {
		w.log.Error("expiry sweep failed", zap.Error(err))
		return
	}
	if n > 0 {
		w.log.Info("expired stale downlinks", zap.Int64("count", n))
	}
}
