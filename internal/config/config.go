package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Provider     ProviderConfig `mapstructure:"provider"`
	DNS          DNSConfig      `mapstructure:"dns"`
	Email        string         `mapstructure:"email" yaml:"email"`
	Organization string         `mapstructure:"organization" yaml:"organization"`
	Owner        string         `mapstructure:"owner" yaml:"owner"`
	Debug        bool
	OutputFormat string
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
	if config.Provider.Location == "" {
		return nil, fmt.Errorf("Location is not set in the config file")
	}
	if config.Provider.Name == "" {
		return nil, fmt.Errorf("Provider is not set in the config file")
	}
	// DEBUG
	fmt.Printf("config: %+v\n", config)
	return &config, nil
}

func setDefaults(v *viper.Viper) {
	// Add any default values here
	// Example:
	// v.SetDefault("some.default.value", "default")
}
