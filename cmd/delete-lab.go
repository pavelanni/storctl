package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewDeleteLabCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lab [name]",
		Short: "Delete a lab",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			labName := args[0]
			assumeYes, _ := cmd.Flags().GetBool("yes")
			skipTimeCheck, _ := cmd.Flags().GetBool("force")

			if !assumeYes && !askForConfirmationSimple("lab", labName) {
				fmt.Println("Operation cancelled")
				return nil
			}

			// Delete the lab using cloud provider
			if status := providerSvc.DeleteLab(labName, skipTimeCheck); status.Error != nil {
				return fmt.Errorf("failed to delete lab: %w", status.Error)
			}

			fmt.Printf("Successfully deleted lab %s\n", labName)
			return nil
		},
	}

	cmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	cmd.Flags().Bool("force", false, "Force deletion without checking DeleteAfter time")
	return cmd
}
