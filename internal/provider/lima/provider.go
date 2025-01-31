package lima

import (
	"fmt"
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
	providerConfig := getProviderConfig(cfg, "lima")
	if providerConfig == nil {
		return nil, fmt.Errorf("provider config not found for lima")
	}

	logger := logger.Get()
	logger.Info("Initializing Lima provider")
	logger.Debug("Using configuration",
		"location", providerConfig.Location)

	arch := runtime.GOARCH
	return &LimaProvider{config: cfg, logger: logger, arch: arch}, nil
}

func (p *LimaProvider) Name() string {
	return "lima"
}

func getProviderConfig(cfg *config.Config, providerName string) *config.ProviderConfig {
	for _, provider := range cfg.Providers {
		if provider.Name == providerName {
			return &provider
		}
	}
	return nil
}
