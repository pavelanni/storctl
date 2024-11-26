package cmd

import "github.com/spf13/cobra"

func NewGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Display information about labs, servers, or volumes",
		Long:  `Get command allows you to retrieve information about various resources like labs, servers, and volumes.`,
	}

	// Add subcommands
	cmd.AddCommand(NewGetLabCmd())
	cmd.AddCommand(NewGetServerCmd())
	cmd.AddCommand(NewGetVolumeCmd())
	cmd.AddCommand(NewGetKeyCmd())

	return cmd
}
