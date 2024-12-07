package cmd

import (
	"fmt"
	"time"

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

			if !assumeYes && !askForConfirmationSimple("volume", volumeName) {
				fmt.Println("Operation cancelled")
				return nil
			}

			// Delete the volume using cloud provider
			status := providerSvc.DeleteVolume(volumeName, skipTimeCheck)
			if status.Error != nil {
				return fmt.Errorf("failed to delete volume: %w", status.Error)
			}
			if !status.Deleted && status.DeleteAfter.After(time.Now().UTC()) {
				fmt.Printf("Volume %s is not ready for deletion until %s UTC\n", volumeName, status.DeleteAfter.Format("2006-01-02 15:04:05"))
				return nil
			}

			fmt.Printf("Successfully deleted volume %s\n", volumeName)
			return nil
		},
	}

	cmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	cmd.Flags().Bool("force", false, "Force deletion without checking DeleteAfter time")
	return cmd
}
