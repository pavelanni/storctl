// Package provider contains the factory for the cloud and local providers.
// It includes the functions to create a new cloud provider based on the configuration.
package provider

import (
	"fmt"

	"github.com/pavelanni/storctl/internal/config"
	"github.com/pavelanni/storctl/internal/provider/hetzner"
	"github.com/pavelanni/storctl/internal/provider/lima"
	"github.com/pavelanni/storctl/internal/provider/virt"
)

// NewProvider creates a new cloud provider based on the configuration
func NewProvider(cfg config.Config, providerName string) (CloudProvider, error) {
	switch providerName {
	case "hetzner":
		return hetzner.New(&cfg)
	case "lima":
		return lima.New(&cfg)
	case "virt":
		return virt.New(&cfg)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", providerName)
	}
}
