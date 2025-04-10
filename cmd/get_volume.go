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

func NewGetVolumeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "volume [volume-id]",
		Short: "Get information about volumes",
		Long:  `Display a list of all volumes or detailed information about a specific volume`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return listVolumes()
			}
			return getVolume(args[0])
		},
	}

	return cmd
}

func listVolumes() error {
	err := initProvider(useProvider)
	if err != nil {
		return err
	}
	volumes, err := providerSvc.AllVolumes()
	if err != nil {
		return err
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tSERVER\tSIZE\tOWNER\tAGE\tDELETE AFTER")
	for _, volume := range volumes {
		deleteAfter := "-"
		if !volume.Status.DeleteAfter.IsZero() {
			deleteAfter = volume.Status.DeleteAfter.Format(time.RFC3339)
		}
		fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%s\t%s\n",
			volume.Name,
			volume.Spec.ServerName,
			volume.Spec.Size,
			volume.Status.Owner,
			timeutil.FormatAge(volume.Status.Created),
			deleteAfter)
	}
	return w.Flush()
}

func getVolume(volumeID string) error {
	err := initProvider(useProvider)
	if err != nil {
		return err
	}
	volume, err := providerSvc.GetVolume(volumeID)
	if err != nil {
		return err
	}

	switch cfg.OutputFormat {
	case "json":
		return output.JSON(volume, os.Stdout)
	case "yaml":
		return output.YAML(volume, os.Stdout)
	default:
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tSERVER\tSIZE\tOWNER\tAGE\tDELETE AFTER")
		deleteAfter := "-"
		if !volume.Status.DeleteAfter.IsZero() {
			deleteAfter = volume.Status.DeleteAfter.Format(time.RFC3339)
		}
		fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%s\t%s\n",
			volume.Name,
			volume.Spec.ServerName,
			volume.Spec.Size,
			volume.Status.Owner,
			timeutil.FormatAge(volume.Status.Created),
			deleteAfter)
		return w.Flush()
	}
}
