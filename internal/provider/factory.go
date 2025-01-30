package provider

import (
	"fmt"

	"github.com/pavelanni/storctl/internal/config"
	"github.com/pavelanni/storctl/internal/provider/hetzner"
	"github.com/pavelanni/storctl/internal/provider/lima"
)

// NewProvider creates a new cloud provider based on the configuration
func NewProvider(cfg config.Config) (CloudProvider, error) {
	switch cfg.Provider.Name {
	case "hetzner":
		return hetzner.New(&cfg)
	case "lima":
		return lima.New(&cfg)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", cfg.Provider.Name)
	}
}
