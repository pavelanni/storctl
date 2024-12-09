package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/pavelanni/storctl/internal/config"
	"github.com/spf13/cobra"
)

func NewConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: fmt.Sprintf("Manage %s configuration", config.ToolName),
		Long:  fmt.Sprintf("View and modify %s configuration settings", config.ToolName),
	}

	cmd.AddCommand(
		newConfigViewCmd(),
		// We'll add other commands later
	)

	return cmd
}

func newConfigViewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "view",
		Short: "View current configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig(cfgFile)
			if err != nil {
				return err
			}

			// Pretty print the config
			output, err := json.MarshalIndent(cfg, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to format config: %w", err)
			}

			fmt.Println(string(output))
			return nil
		},
	}
}
