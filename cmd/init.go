package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pavelanni/labshop/internal/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func NewInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize labshop",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := createConfig(); err != nil {
				return fmt.Errorf("error creating config: %w", err)
			}
			if err := createTemplates(); err != nil {
				return fmt.Errorf("error creating templates: %w", err)
			}
			return nil
		},
	}
}

func createConfig() error {
	// Create the default config directory if it doesn't exist and copy the default config file
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("error getting home directory: %w", err)
	}
	configDir := filepath.Join(home, config.DefaultConfigDir)
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		err = os.MkdirAll(configDir, 0755)
		if err != nil {
			return fmt.Errorf("error creating config directory: %w", err)
		}
	}
	cfgFile := filepath.Join(configDir, config.ConfigFileName)
	// if the config file already exists, print a message and exit
	if _, err := os.Stat(cfgFile); !os.IsNotExist(err) {
		fmt.Printf("Config file already exists at %s\n", cfgFile)
		return nil
	}
	var defaultCfg config.Config
	defaultCfg.Provider.Name = config.DefaultProvider
	defaultCfg.Provider.Location = config.DefaultLocation
	defaultCfg.Provider.Token = config.DefaultToken
	defaultCfg.Provider.Credentials = map[string]string{
		"username": config.DefaultCredentials,
		"password": config.DefaultCredentials,
	}
	defaultCfg.DNS.Provider = config.DefaultDNSProvider
	defaultCfg.DNS.ZoneID = config.DefaultDNSZoneID
	defaultCfg.DNS.Token = config.DefaultDNSToken
	defaultCfg.DNS.Domain = config.DefaultDomain
	defaultCfg.Email = config.DefaultEmail
	defaultCfg.Organization = config.DefaultOrganization
	defaultCfg.Owner = config.DefaultOwner

	// Marshal the default config to YAML and write it to the default config file
	cfgBytes, err := yaml.Marshal(defaultCfg)
	if err != nil {
		return fmt.Errorf("error marshalling default config: %w", err)
	}
	err = os.WriteFile(cfgFile, cfgBytes, 0600)
	if err != nil {
		return fmt.Errorf("error writing default config: %w", err)
	}
	fmt.Printf("Config file created at %s\n", cfgFile)
	return nil
}

func createTemplates() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("error getting home directory: %w", err)
	}
	configDir := filepath.Join(home, config.DefaultConfigDir)
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		err = os.MkdirAll(configDir, 0755)
		if err != nil {
			return fmt.Errorf("error creating config directory: %w", err)
		}
	}
	// Create the default templates directory if it doesn't exist and copy the default template file
	templatesDir := filepath.Join(configDir, config.DefaultTemplateDir)
	if _, err := os.Stat(templatesDir); os.IsNotExist(err) {
		err = os.MkdirAll(templatesDir, 0755)
		if err != nil {
			return fmt.Errorf("error creating templates directory: %w", err)
		}
	}
	labTemplateFile := filepath.Join(templatesDir, "lab.yaml")
	// if the lab template file already exists, print a message and exit
	if _, err := os.Stat(labTemplateFile); !os.IsNotExist(err) {
		fmt.Printf("Lab template file already exists at %s\n", labTemplateFile)
		return nil
	}
	err = os.WriteFile(labTemplateFile, []byte(config.DefaultLabTemplate), 0644)
	if err != nil {
		return fmt.Errorf("error writing lab template file: %w", err)
	}
	fmt.Printf("Lab template file created at %s\n", labTemplateFile)
	return nil
}