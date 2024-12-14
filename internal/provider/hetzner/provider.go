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
	token := cfg.Provider.Token
	if token == "" {
		return nil, fmt.Errorf("Hetzner API token is required")
	}

	// Create a new logger with the configured log level
	logLevel := logger.ParseLevel(cfg.LogLevel)
	providerLogger := logger.NewLogger(logLevel)

	client := hcloud.NewClient(hcloud.WithToken(token))
	p := &HetznerProvider{
		Client: client,
		config: cfg,
		logger: providerLogger,
	}

	return p, nil
}
