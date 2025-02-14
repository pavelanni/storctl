// Package cmd contains all commands for the storctl tool.
// It includes commands to get, create, delete, and sync lab resources.
package cmd

import (
	"fmt"
	"os"
	"os/exec"
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
	useProvider string
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
			// Check if prerequisites are met
			if err := checkPrerequisites(useProvider); err != nil {
				fmt.Fprintf(os.Stderr, "Prerequisites not met: %v\n", err)
				os.Exit(1)
			}
			// Continue with other initializations
			initConfig()
			initDNS()
			return nil
		},
	}

	// Global flags
	defaultConfigFile := filepath.Join(os.Getenv("HOME"), config.DefaultConfigDir, "config.yaml")
	cmd.PersistentFlags().StringVar(&cfgFile, "config", defaultConfigFile, "config file")
	cmd.PersistentFlags().StringVar(&logLevel, "log-level", "warn", "logging level (debug, info, warn, error)")
	cmd.PersistentFlags().StringVar(&useProvider, "provider", config.DefaultLocalProvider, "Provider to use")
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
		NewVersionCmd(),
		NewInstallCmd(),
	)

	return cmd
}

func checkPrerequisites(provider string) error {
	// Check if kubectl is installed
	if _, err := exec.LookPath("kubectl"); err != nil {
		return fmt.Errorf("kubectl is not installed. Please follow the instructions at https://kubernetes.io/docs/tasks/tools/#kubectl")
	}

	// Check if helm is installed
	if _, err := exec.LookPath("helm"); err != nil {
		return fmt.Errorf("helm is not installed. Please follow the instructions at https://helm.sh/docs/intro/install/")
	}

	// Check if krew is installed
	if _, err := exec.LookPath("kubectl-krew"); err != nil {
		return fmt.Errorf("krew is not installed. Please follow the instructions at https://krew.sigs.k8s.io/docs/user-guide/setup/install/")
	}

	// Check if directpv is installed
	if _, err := exec.LookPath("kubectl-directpv"); err != nil {
		return fmt.Errorf("directpv is not installed. Please follow the instructions at https://min.io/docs/directpv/installation/#install-directpv-plugin-with-krew")
	}

	if provider == "lima" {
		// Check if lima is installed
		if _, err := exec.LookPath("limactl"); err != nil {
			return fmt.Errorf("lima is not installed. Please follow the instructions at https://lima-vm.io/docs/installation/")
		}
		//Check if socket_vmnet is installed
		if _, err := os.Stat("/opt/socket_vmnet/bin/socket_vmnet"); err != nil {
			return fmt.Errorf("socket_vmnet is not installed. Please follow the instructions at https://lima-vm.io/docs/config/network/#socket_vmnet")
		}
		// Check if sudoers file is present
		if _, err := os.Stat("/etc/sudoers.d/lima"); err != nil {
			return fmt.Errorf("sudoers file for Lima is not present. Please follow the instructions at https://lima-vm.io/docs/config/network/#socket_vmnet")
		}
	}

	return nil
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

	viper.SetDefault("storage.path", filepath.Join(os.Getenv("HOME"), config.DefaultConfigDir, config.DefaultLabStorageFile))
	viper.SetDefault("storage.bucket", config.DefaultLabBucket)

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

func initProvider(providerName string) error {
	var err error

	providerSvc, err = provider.NewProvider(*cfg, providerName)
	if err != nil {
		return fmt.Errorf("error initializing provider: %w", err)
	}
	return nil
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

func initLabManager() error {
	var err error
	labSvc, err = lab.NewManager(providerSvc, cfg)
	if err != nil {
		return fmt.Errorf("error initializing lab manager: %w", err)
	}
	return nil
}

func Execute() error {
	return NewRootCmd().Execute()
}
