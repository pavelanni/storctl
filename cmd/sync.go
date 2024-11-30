package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewSyncCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "Sync labs",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := providerSvc.SyncLabs(); err != nil {
				return fmt.Errorf("error syncing labs: %w", err)
			}
			return nil
		},
	}
}
