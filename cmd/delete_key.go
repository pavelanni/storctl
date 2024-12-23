package cmd

import (
	"fmt"
	"time"

	"github.com/pavelanni/storctl/internal/ssh"
	"github.com/spf13/cobra"
)

func NewDeleteSSHKeyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "key [name]",
		Short: "Delete an SSH key",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			keyName := args[0]
			force, _ := cmd.Flags().GetBool("yes")
			skipTimeCheck, _ := cmd.Flags().GetBool("force")

			if !force && !askForConfirmationSimple("key", keyName) {
				fmt.Println("Operation cancelled")
				return nil
			}

			// Delete the key using cloud provider
			status := providerSvc.DeleteSSHKey(keyName, skipTimeCheck)
			if status.Error != nil {
				return fmt.Errorf("failed to delete key: %w", status.Error)
			}
			if !status.Deleted && status.DeleteAfter.After(time.Now().UTC()) {
				fmt.Printf("Key %s is not ready for deletion until %s UTC\n", keyName, status.DeleteAfter.Format("2006-01-02 15:04:05"))
				return nil
			}
			// delete the local key pair
			keyManager := ssh.NewManager(cfg)
			err := keyManager.DeleteLocalKeyPair(keyName)
			if err != nil {
				return fmt.Errorf("failed to delete local key pair: %w", err)
			}
			fmt.Printf("Successfully deleted key %s\n", keyName)
			return nil
		},
	}

	cmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	cmd.Flags().Bool("force", false, "Force deletion without checking DeleteAfter time")
	return cmd
}
