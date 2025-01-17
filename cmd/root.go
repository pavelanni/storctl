package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pavelanni/storctl/internal/config"
	"github.com/pavelanni/storctl/internal/dns"
	"github.com/pavelanni/storctl/internal/lab"
	"github.com/pavelanni/storctl/internal/logger"
	"github.com/pavelanni/storctl/internal/provider"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile     string
	cfg         *config.Config
	providerSvc provider.CloudProvider
	dnsSvc      *dns.CloudflareDNSProvider
	labSvc      *lab.ManagerSvc
	logLevel    string
)

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   config.ToolName,
		Short: fmt.Sprintf("%s - AIStor Environment Manager", config.ToolName),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Initialize logger first
			logLevel := logger.ParseLevel(viper.GetString("log_level"))
			logger.Initialize(logLevel)

			// Skip other initializations for init command
			if cmd.Name() == "init" {
				return nil
			}

			// Continue with other initializations
			initConfig()
			initProvider()
			initDNS()
			initLabManager()
			return nil
		},
	}

	// Global flags
	defaultConfigFile := filepath.Join(os.Getenv("HOME"), config.DefaultConfigDir, "config.yaml")
	cmd.PersistentFlags().StringVar(&cfgFile, "config", defaultConfigFile, "config file")
	cmd.PersistentFlags().StringVar(&logLevel, "log-level", "warn", "logging level (debug, info, warn, error)")
	err := viper.BindPFlag("log_level", cmd.PersistentFlags().Lookup("log-level"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error binding log level flag: %v\n", err)
		os.Exit(1)
	}

	// Add commands
	cmd.AddCommand(NewInitCmd(),
		NewGetCmd(),
		NewDeleteCmd(),
		NewConfigCmd(),
		NewCreateCmd(),
		NewSyncCmd(),
	)

	return cmd
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting home directory: %v\n", err)
			os.Exit(1)
		}

		configDir := filepath.Join(home, config.DefaultConfigDir)
		if err := os.MkdirAll(configDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating config directory: %v\n", err)
			os.Exit(1)
		}

		viper.AddConfigPath(configDir)
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	viper.AutomaticEnv()
	viper.SetEnvPrefix(strings.ToUpper(config.ToolName))

	viper.SetDefault("storage.path", filepath.Join(os.Getenv("HOME"), config.DefaultConfigDir, "labs.db"))
	viper.SetDefault("storage.bucket", "labs")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Fprintf(os.Stderr, "Error reading config: %v\n", err)
			os.Exit(1)
		}
	}

	cfg = &config.Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error unmarshaling config: %v\n", err)
		os.Exit(1)
	}

	cfg.LogLevel = viper.GetString("log_level")
}

func initProvider() {
	var err error

	providerSvc, err = provider.NewProvider(*cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing provider: %v\n", err)
		os.Exit(1)
	}
}

func initDNS() {
	var err error

	if cfg.DNS.Provider == "cloudflare" {
		dnsSvc, err = dns.NewCloudflareDNS(cfg.DNS.Token)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing DNS provider: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Fprintf(os.Stderr, "Unsupported DNS provider: %s\n", cfg.DNS.Provider)
		os.Exit(1)
	}
}

func initLabManager() {
	var err error
	labSvc, err = lab.NewManager(providerSvc, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing lab manager: %v\n", err)
		os.Exit(1)
	}
}

func Execute() error {
	return NewRootCmd().Execute()
}
