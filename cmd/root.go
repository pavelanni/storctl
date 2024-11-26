package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pavelanni/labshop/internal/config"
	"github.com/pavelanni/labshop/internal/provider"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile     string
	cfg         *config.Config
	providerSvc provider.CloudProvider
)

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "labshop",
		Short: "Labshop - Lab Environment Manager",
	}

	// Global flags
	cmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.labshop/config.yaml)")
	cmd.PersistentFlags().Bool("debug", false, "enable debug output")

	// Add commands
	cmd.AddCommand(
		NewGetCmd(),
		NewDeleteCmd(),
		NewConfigCmd(),
		NewCreateCmd(),
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

		configDir := filepath.Join(home, ".labshop")
		if err := os.MkdirAll(configDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating config directory: %v\n", err)
			os.Exit(1)
		}

		viper.AddConfigPath(configDir)
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	viper.AutomaticEnv()
	viper.SetEnvPrefix("LABSHOP")

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
}

func initProvider() {
	var err error

	providerSvc, err = provider.NewProvider(*cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing provider: %v\n", err)
		os.Exit(1)
	}
}

func Execute() error {
	return NewRootCmd().Execute()
}

func init() {
	cobra.OnInitialize(initConfig, initProvider)
}
