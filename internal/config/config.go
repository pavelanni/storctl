// Package config contains the configuration for the storctl tool.
// It includes the configuration for the providers, DNS, storage, email, organization, owner, output format, log level, and ansible.
// The configuration is read from a YAML file and can be overridden by environment variables.
// This package also contains the default values for the configuration in the constants.go file.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	Providers    []ProviderConfig `mapstructure:"providers"`
	DNS          DNSConfig        `mapstructure:"dns"`
	Storage      StorageConfig    `mapstructure:"storage"`
	Email        string           `mapstructure:"email" yaml:"email"`
	Organization string           `mapstructure:"organization" yaml:"organization"`
	Owner        string           `mapstructure:"owner" yaml:"owner"`
	OutputFormat string           `mapstructure:"output_format" yaml:"output_format"`
	LogLevel     string           `mapstructure:"log_level" yaml:"log_level"`
	Ansible      AnsibleConfig    `mapstructure:"ansible" yaml:"ansible"`
}

type StorageConfig struct {
	Path   string `mapstructure:"path" yaml:"path"`
	Bucket string `mapstructure:"bucket" yaml:"bucket"`
}

type ProviderConfig struct {
	Name        string            `mapstructure:"name"`
	Location    string            `mapstructure:"location"`
	Token       string            `mapstructure:"token"`
	Credentials map[string]string `mapstructure:"credentials"`
}

type DNSConfig struct {
	Provider string `mapstructure:"provider"`
	Domain   string `mapstructure:"domain"`
	Token    string `mapstructure:"token"`
	ZoneID   string `mapstructure:"zone_id"`
}

type AnsibleConfig struct {
	ConfigFile string `mapstructure:"config_file"`
}

// LoadConfig reads configuration from file and environment variables
func LoadConfig(configPath string) (*Config, error) {
	v := viper.New()

	// Set defaults if any
	setDefaults(v)

	// Configure Viper to read from file
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	// Configure environment variables
	v.AutomaticEnv()
	// Map environment variables to config fields
	if err := v.BindEnv("provider.token", "PROVIDER_TOKEN"); err != nil {
		return nil, fmt.Errorf("failed to bind PROVIDER_TOKEN: %w", err)
	}
	if err := v.BindEnv("dns.token", "DNS_TOKEN"); err != nil {
		return nil, fmt.Errorf("failed to bind DNS_TOKEN: %w", err)
	}
	if err := v.BindEnv("dns.zone_id", "DNS_ZONE_ID"); err != nil {
		return nil, fmt.Errorf("failed to bind DNS_ZONE_ID: %w", err)
	}

	// Read the config file
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Unmarshal config into struct
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	if config.Owner == "" {
		return nil, fmt.Errorf("Owner is not set in the config file")
	}
	if config.Organization == "" {
		return nil, fmt.Errorf("Organization is not set in the config file")
	}
	if config.Email == "" {
		return nil, fmt.Errorf("Email is not set in the config file")
	}
	if len(config.Providers) == 0 {
		return nil, fmt.Errorf("Providers are not set in the config file. You should have at least one provider")
	}
	return &config, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("storage.path", filepath.Join(os.Getenv("HOME"), DefaultConfigDir, DefaultLabStorageFile))
	v.SetDefault("storage.bucket", DefaultLabBucket)
}
