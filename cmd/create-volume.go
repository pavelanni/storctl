package cmd

import (
	"fmt"

	"github.com/pavelanni/labshop/internal/types"
	"github.com/spf13/cobra"
)

func NewCreateVolumeCmd() *cobra.Command {
	var (
		size      int
		server    string
		labels    map[string]string
		automount bool
		format    string
	)

	cmd := &cobra.Command{
		Use:   "volume [name]",
		Short: "Create a new volume",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			volumeName := args[0]
			_, err := providerSvc.CreateVolume(volumeName, size, server, labels, automount, format)
			return err
		},
	}

	cmd.Flags().IntVar(&size, "size", 10, "Volume size in GB")
	cmd.Flags().StringVar(&server, "server", "", "Server to attach the volume to")
	cmd.Flags().StringToStringVar(&labels, "labels", map[string]string{}, "Volume labels")
	cmd.Flags().BoolVar(&automount, "automount", false, "Automount the volume")
	cmd.Flags().StringVar(&format, "format", "xfs", "Volume format")
	if err := cmd.MarkFlagRequired("server"); err != nil {
		panic(err)
	}

	return cmd
}

func createVolume(volume *types.Resource) error {
	fmt.Printf("Creating volume %s with size %d\n",
		volume.Metadata["name"],
		volume.Spec["size"])
	if volume.Spec["server"] != nil {
		fmt.Printf("  server: %s\n", volume.Spec["server"])
	}
	if volume.Spec["automount"] != nil {
		fmt.Printf("  automount: %t\n", volume.Spec["automount"])
	}
	if volume.Spec["format"] != nil {
		fmt.Printf("  format: %s\n", volume.Spec["format"])
	}
	labels := make(map[string]string)
	for k, v := range volume.Metadata["labels"].(map[string]interface{}) {
		labels[k] = v.(string)
	}
	_, err := providerSvc.CreateVolume(
		volume.Metadata["name"].(string),
		int(volume.Spec["size"].(float64)),
		volume.Spec["server"].(string),
		labels,
		volume.Spec["automount"].(bool),
		volume.Spec["format"].(string),
	)
	return err
}
