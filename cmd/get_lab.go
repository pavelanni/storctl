package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/pavelanni/storctl/internal/util/output"
	"github.com/pavelanni/storctl/internal/util/timeutil"
	"github.com/spf13/cobra"
)

func NewGetLabCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lab [lab-id]",
		Short: "Get information about labs",
		Long:  `Display a list of all active labs or detailed information about a specific lab`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return listLabs()
			}
			return getLab(args[0])
		},
	}

	return cmd
}

func listLabs() error {
	err := initProvider(useProvider)
	if err != nil {
		return err
	}
	err = initLabManager()
	if err != nil {
		return err
	}
	labs, err := labSvc.List()
	if err != nil {
		return err
	}
	// Create a new tabwriter
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	// Print header
	_, err = fmt.Fprintln(w, "NAME\tOWNER\tNODES\tTYPE\tVOLS\tSIZE\tAGE\tDELETE-AFTER")
	if err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Print data for each lab
	for _, lab := range labs {
		serverType := "N/A"
		volSize := 0
		owner := "N/A"
		labAge := "N/A"
		deleteAfter := time.Time{}
		if len(lab.Status.Servers) > 0 {
			serverType = lab.Status.Servers[0].Spec.ServerType
			deleteAfter = lab.Status.Servers[0].Status.DeleteAfter
			if lab.Status.Servers[0].Status.Owner != "" {
				owner = lab.Status.Servers[0].Status.Owner
			}
			labAge = timeutil.FormatAge(lab.Status.Servers[0].Status.Created)
		}
		if len(lab.Status.Volumes) > 0 {
			volSize = lab.Status.Volumes[0].Spec.Size
		}
		deleteAfterStr := "N/A"
		if !deleteAfter.IsZero() {
			deleteAfterStr = deleteAfter.Format(time.RFC3339)
		}

		_, err = fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%d\t%.2f\t%s\t%s\n",
			lab.Name, owner, len(lab.Status.Servers), serverType, len(lab.Status.Volumes), float32(volSize), labAge, deleteAfterStr)
		if err != nil {
			return fmt.Errorf("failed to write lab: %w", err)
		}
	}

	// Flush the tabwriter to output
	err = w.Flush()
	if err != nil {
		return fmt.Errorf("failed to flush writer: %w", err)
	}
	return nil
}

func getLab(labName string) error {
	err := initProvider(useProvider)
	if err != nil {
		return err
	}
	err = initLabManager()
	if err != nil {
		return err
	}
	lab, err := labSvc.Get(labName)
	if err != nil {
		return err
	}
	switch cfg.OutputFormat {
	case "json":
		return output.JSON(lab, os.Stdout)
	case "yaml":
		return output.YAML(lab, os.Stdout)
	default:
		fmt.Printf("Lab: %s\n", lab.Name)
		for _, server := range lab.Status.Servers {
			fmt.Printf("  Server: %s, Type: %s, Cores: %d, Memory: %.2fGB, Disk: %dGB, DeleteAfter: %s\n",
				server.Name,
				server.Spec.ServerType,
				server.Status.Cores,
				server.Status.Memory,
				server.Status.Disk,
				server.Status.DeleteAfter)
		}
		for _, volume := range lab.Status.Volumes {
			fmt.Printf("  Volume: %s, Size: %dGB, DeleteAfter: %s\n",
				volume.Name,
				volume.Spec.Size,
				volume.Status.DeleteAfter)
		}
	}

	return nil
}
