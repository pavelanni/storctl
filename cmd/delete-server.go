package cmd

import (
	"fmt"
	"time"

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

			if !assumeYes && !askForConfirmationSimple("server", serverName) {
				fmt.Println("Operation cancelled")
				return nil
			}

			// Delete the server using cloud provider
			status := providerSvc.DeleteServer(serverName, skipTimeCheck)
			if status.Error != nil {
				return fmt.Errorf("failed to delete server: %w", status.Error)
			}
			if !status.Deleted && status.DeleteAfter.After(time.Now().UTC()) {
				fmt.Printf("Server %s is not ready for deletion until %s UTC\n", serverName, status.DeleteAfter.Format("2006-01-02 15:04:05"))
				return nil
			}

			fmt.Printf("Successfully deleted server %s\n", serverName)
			return nil
		},
	}

	cmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	cmd.Flags().Bool("force", false, "Force deletion without checking DeleteAfter time")
	return cmd
}
