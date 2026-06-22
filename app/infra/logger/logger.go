package logger

import (
	"github.com/anoop-dryad/bridgehead/app/config"

	"go.uber.org/zap"
)

func New(cfg config.App) (*zap.Logger, error) {
	if cfg.IsProduction {
		return zap.NewProduction() // Error+ only, JSON
	}
	return zap.NewDevelopment() // Debug+, human readable
}
