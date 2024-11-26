package hetzner

import (
	"fmt"
	"log/slog"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/pavelanni/labshop/internal/config"
	"github.com/pavelanni/labshop/internal/logger"
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

	client := hcloud.NewClient(hcloud.WithToken(token))
	return &HetznerProvider{
		Client: client,
		config: cfg,
		logger: logger.GetLogger(),
	}, nil
}
