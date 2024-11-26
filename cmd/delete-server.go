package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewDeleteServerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "server [name]",
		Short: "Delete a server",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			serverName := args[0]
			assumeYes, _ := cmd.Flags().GetBool("yes")
			skipTimeCheck, _ := cmd.Flags().GetBool("force")

			if !assumeYes {
				fmt.Printf("Are you sure you want to delete server %s? [y/N] ", serverName)
				var response string
				_, err := fmt.Scanln(&response)
				if err != nil {
					return fmt.Errorf("failed to read response: %w", err)
				}
				if response != "y" && response != "Y" {
					fmt.Println("Operation cancelled")
					return nil
				}
			}

			// Delete the server using cloud provider
			if err := providerSvc.DeleteServer(serverName, skipTimeCheck); err != nil {
				return fmt.Errorf("failed to delete server: %w", err)
			}

			fmt.Printf("Successfully deleted server %s\n", serverName)
			return nil
		},
	}

	cmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	cmd.Flags().Bool("force", false, "Force deletion without checking DeleteAfter time")
	return cmd
}
