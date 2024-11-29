package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/pavelanni/labshop/internal/util/timeutil"
	"github.com/spf13/cobra"
)

func NewGetServerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "server [server-id]",
		Short: "Get information about servers",
		Long:  `Display a list of all active servers or detailed information about a specific server`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return listServers()
			}
			return getServer(args[0])
		},
	}

	return cmd
}

func listServers() error {
	servers, err := providerSvc.AllServers()
	if err != nil {
		return err
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tTYPE\tOWNER\tAGE\tDELETE AFTER")
	for _, server := range servers {
		deleteAfter := "-"
		if !server.Status.DeleteAfter.IsZero() {
			deleteAfter = server.Status.DeleteAfter.Format(time.RFC3339)
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			server.Name,
			server.Spec.Type,
			server.Status.Owner,
			timeutil.FormatAge(server.Status.Created),
			deleteAfter)
	}
	return w.Flush()
}

func getServer(serverID string) error {
	server, err := providerSvc.GetServer(serverID)
	if err != nil {
		return err
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tTYPE\tOWNER\tAGE\tDELETE AFTER")
	deleteAfter := "-"
	if !server.Status.DeleteAfter.IsZero() {
		deleteAfter = server.Status.DeleteAfter.Format(time.RFC3339)
	}
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
		server.Name,
		server.Spec.Type,
		server.Status.Owner,
		timeutil.FormatAge(server.Status.Created),
		deleteAfter)
	return w.Flush()
}
