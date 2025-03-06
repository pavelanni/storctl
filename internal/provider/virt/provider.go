// Package virt contains the libvirt implementation of the provider interface for the storctl tool.
// It includes the functions to create, get, list, and delete volumes.
// It also includes the functions to get the server and key for the lab.
package virt

import (
	"fmt"
	"log/slog"
	"net/url"
	"runtime"

	"github.com/digitalocean/go-libvirt"
	"github.com/pavelanni/storctl/internal/config"
	"github.com/pavelanni/storctl/internal/logger"
)

type VirtProvider struct {
	config *config.Config
	logger *slog.Logger
	arch   string
	client *libvirt.Libvirt
}

func New(cfg *config.Config) (*VirtProvider, error) {
	providerConfig := getProviderConfig(cfg, "virt")
	if providerConfig == nil {
		return nil, fmt.Errorf("provider config not found for virt")
	}

	logger := logger.Get()
	logger.Info("Initializing Virt provider")
	logger.Debug("Using configuration",
		"location", providerConfig.Location)

	arch := runtime.GOARCH
	uri, _ := url.Parse(string(libvirt.QEMUSystem))
	client, err := libvirt.ConnectToURI(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to libvirt: %w", err)
	}
	return &VirtProvider{config: cfg, logger: logger, arch: arch, client: client}, nil
}

func (p *VirtProvider) Name() string {
	return "virt"
}

func getProviderConfig(cfg *config.Config, providerName string) *config.ProviderConfig {
	for _, provider := range cfg.Providers {
		if provider.Name == providerName {
			return &provider
		}
	}
	return nil
}
