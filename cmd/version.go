package cmd

import (
	"fmt"
	"runtime"

	"github.com/pavelanni/storctl/internal/version"
	"github.com/spf13/cobra"
)

func NewVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Version: %s\n", version.Version)
			fmt.Printf("Commit: %s\n", version.Commit)
			fmt.Printf("Build Date: %s\n", version.Date)
			fmt.Printf("Go Version: %s\n", version.GoVersion)
			fmt.Printf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
		},
	}
}
