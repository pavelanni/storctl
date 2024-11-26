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

			if !assumeYes {
				fmt.Printf("Are you sure you want to delete lab %s? [y/N] ", labName)
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

			// Delete the lab using cloud provider
			if err := providerSvc.DeleteLab(labName, skipTimeCheck); err != nil {
				return fmt.Errorf("failed to delete lab: %w", err)
			}

			fmt.Printf("Successfully deleted lab %s\n", labName)
			return nil
		},
	}

	cmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	cmd.Flags().Bool("force", false, "Force deletion without checking DeleteAfter time")
	return cmd
}
