package lima

import (
	"log/slog"
	"runtime"

	"github.com/pavelanni/storctl/internal/config"
	"github.com/pavelanni/storctl/internal/logger"
)

type LimaProvider struct {
	config *config.Config
	logger *slog.Logger
	arch   string
}

func New(cfg *config.Config) (*LimaProvider, error) {
	logger := logger.Get()
	logger.Info("Initializing Lima provider")
	logger.Debug("Using configuration",
		"location", cfg.Provider.Location)

	arch := runtime.GOARCH
	return &LimaProvider{config: cfg, logger: logger, arch: arch}, nil
}
