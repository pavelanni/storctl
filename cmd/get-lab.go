package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/pavelanni/labshop/internal/types"
	"github.com/pavelanni/labshop/internal/util/timeutil"
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
	labs := make(map[string]*types.Lab)
	servers, err := providerSvc.AllServers()
	if err != nil {
		return err
	}
	volumes, err := providerSvc.AllVolumes()
	if err != nil {
		return err
	}
	for _, server := range servers {
		labName := server.Labels["lab_name"]
		if labName == "" {
			continue
		}
		if labs[labName] == nil {
			labs[labName] = &types.Lab{
				TypeMeta: types.TypeMeta{
					APIVersion: "v1",
					Kind:       "Lab",
				},
				ObjectMeta: types.ObjectMeta{
					Name: labName,
				},
			}
		}
		labs[labName].Status.Servers = append(labs[labName].Status.Servers, server)
	}
	for _, volume := range volumes {
		labName := volume.Labels["lab_name"]
		if labName == "" {
			continue
		}
		if labs[labName] == nil {
			labs[labName] = &types.Lab{
				TypeMeta: types.TypeMeta{
					APIVersion: "v1",
					Kind:       "Lab",
				},
				ObjectMeta: types.ObjectMeta{
					Name: labName,
				},
			}
		}
		labs[labName].Status.Volumes = append(labs[labName].Status.Volumes, volume)
	}

	// Create a new tabwriter
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	// Print header
	fmt.Fprintln(w, "NAME\tOWNER\tNODES\tTYPE\tVOLS\tSIZE\tAGE\tDELETE-AFTER")

	// Print data for each lab
	for _, lab := range labs {
		serverType := "N/A"
		volSize := 0
		owner := "N/A"
		labAge := "N/A"
		deleteAfter := time.Time{}
		if len(lab.Status.Servers) > 0 {
			serverType = lab.Status.Servers[0].Spec.Type
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

		fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%d\t%.2f\t%s\t%s\n",
			lab.Name, owner, len(lab.Status.Servers), serverType, len(lab.Status.Volumes), float32(volSize), labAge, deleteAfterStr)
	}

	// Flush the tabwriter to output
	w.Flush()
	return nil
}

func getLab(labName string) error {
	fmt.Printf("Getting details for lab: %s\n", labName)
	lab, err := providerSvc.GetLab(labName)
	if err != nil {
		return err
	}
	fmt.Printf("Lab: %s\n", lab.Name)
	for _, server := range lab.Status.Servers {
		fmt.Printf("  Server: %s, Type: %s, Cores: %d, Memory: %.2fGB, Disk: %dGB, DeleteAfter: %s\n",
			server.ObjectMeta.Name,
			server.Spec.Type,
			server.Status.Cores,
			server.Status.Memory,
			server.Status.Disk,
			server.Status.DeleteAfter)
	}
	for _, volume := range lab.Status.Volumes {
		fmt.Printf("  Volume: %s, Size: %dGB, DeleteAfter: %s\n", volume.ObjectMeta.Name, volume.Spec.Size, volume.Status.DeleteAfter)
	}

	return nil
}