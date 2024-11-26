package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pavelanni/labshop/internal/config"
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

			if !force {
				fmt.Printf("Are you sure you want to delete key %s? [y/N] ", keyName)
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

			// Delete the key using cloud provider
			if err := providerSvc.DeleteSSHKey(keyName, skipTimeCheck); err != nil {
				return fmt.Errorf("failed to delete key: %w", err)
			}
			// Delete the key from the ~/.labshop/keys directory
			if err := os.Remove(filepath.Join(os.Getenv("HOME"), config.DefaultConfigDir, config.KeysDir, keyName)); err != nil {
				return fmt.Errorf("failed to delete private key from the keys directory: %w", err)
			}
			// Delete the public key from the ~/.labshop/keys directory
			if err := os.Remove(filepath.Join(os.Getenv("HOME"), config.DefaultConfigDir, config.KeysDir, keyName+".pub")); err != nil {
				return fmt.Errorf("failed to delete public key from the keys directory: %w", err)
			}

			fmt.Printf("Successfully deleted key %s\n", keyName)
			return nil
		},
	}

	cmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	cmd.Flags().Bool("force", false, "Force deletion without checking DeleteAfter time")
	return cmd
}
