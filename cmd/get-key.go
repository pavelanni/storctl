package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/pavelanni/labshop/internal/util/timeutil"
	"github.com/spf13/cobra"
)

func NewGetKeyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "key [name]",
		Short: "List SSH keys",
		Long:  `Display information about SSH keys including lab name, age, and deletion time.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return listKeys()
			}
			return getKey(args[0])
		},
	}

	return cmd
}

func listKeys() error {
	keys, err := providerSvc.AllSSHKeys()
	if err != nil {
		return err
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tLAB\tAGE\tDELETE AFTER")
	for _, key := range keys {
		deleteAfter := "-"
		if !key.DeleteAfter.IsZero() {
			deleteAfter = key.DeleteAfter.Format(time.RFC3339)
		}

		labName := key.Labels["lab_name"]
		if labName == "" {
			labName = "-"
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			key.Name,
			labName,
			timeutil.FormatAge(key.Created),
			deleteAfter)
	}
	return w.Flush()
}

func getKey(name string) error {
	key, err := providerSvc.GetSSHKey(name)
	if err != nil {
		return err
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tLAB\tAGE\tDELETE AFTER")

	deleteAfter := "-"
	if !key.DeleteAfter.IsZero() {
		deleteAfter = key.DeleteAfter.Format(time.RFC3339)
	}

	labName := key.Labels["lab_name"]
	if labName == "" {
		labName = "-"
	}

	fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
		key.Name,
		labName,
		timeutil.FormatAge(key.Created),
		deleteAfter)

	return w.Flush()
}
