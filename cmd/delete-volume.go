package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewDeleteVolumeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "volume [name]",
		Short: "Delete a volume",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			volumeName := args[0]
			assumeYes, _ := cmd.Flags().GetBool("yes")
			skipTimeCheck, _ := cmd.Flags().GetBool("force")

			if !assumeYes {
				fmt.Printf("Are you sure you want to delete volume %s? [y/N] ", volumeName)
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

			// Delete the volume using cloud provider
			if err := providerSvc.DeleteVolume(volumeName, skipTimeCheck); err != nil {
				return fmt.Errorf("failed to delete volume: %w", err)
			}

			fmt.Printf("Successfully deleted volume %s\n", volumeName)
			return nil
		},
	}

	cmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	cmd.Flags().Bool("force", false, "Force deletion without checking DeleteAfter time")
	return cmd
}
