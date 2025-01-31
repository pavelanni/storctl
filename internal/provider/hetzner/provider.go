package hetzner

import (
	"fmt"
	"log/slog"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/pavelanni/storctl/internal/config"
	"github.com/pavelanni/storctl/internal/logger"
)

type HetznerProvider struct {
	Client *hcloud.Client
	config *config.Config
	logger *slog.Logger
}

func New(cfg *config.Config) (*HetznerProvider, error) {
	providerConfig := getProviderConfig(cfg, "hetzner")
	if providerConfig == nil {
		return nil, fmt.Errorf("provider config not found for hetzner")
	}

	token := providerConfig.Token
	if token == "" {
		return nil, fmt.Errorf("Hetzner API token is required")
	}

	// Create a new logger with the configured log level
	logger := logger.Get()
	logger.Info("Initializing Hetzner provider")
	logger.Debug("Using configuration",
		"location", providerConfig.Location,
		"credentials_present", providerConfig.Token != "")

	client := hcloud.NewClient(hcloud.WithToken(token))
	p := &HetznerProvider{
		Client: client,
		config: cfg,
		logger: logger,
	}

	return p, nil
}

func (p *HetznerProvider) Name() string {
	return "hetzner"
}

func getProviderConfig(cfg *config.Config, providerName string) *config.ProviderConfig {
	for _, provider := range cfg.Providers {
		if provider.Name == providerName {
			return &provider
		}
	}
	return nil
}
