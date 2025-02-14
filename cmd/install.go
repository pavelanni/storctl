package cmd

import "github.com/spf13/cobra"

func NewInstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "install",
		Short:                 "Install software in the environment",
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Show help message if no subcommand is provided
			return cmd.Help()
		},
	}

	cmd.AddCommand(
		NewInstallLabCmd(),
	)

	return cmd
}
