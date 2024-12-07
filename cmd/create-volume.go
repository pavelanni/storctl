package cmd

import (
	"fmt"

	"github.com/pavelanni/labshop/internal/config"
	"github.com/pavelanni/labshop/internal/provider/options"
	"github.com/pavelanni/labshop/internal/types"
	"github.com/pavelanni/labshop/internal/util/labelutil"
	"github.com/pavelanni/labshop/internal/util/timeutil"
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
			volume := &types.Volume{
				TypeMeta: types.TypeMeta{
					Kind:       "Volume",
					APIVersion: "v1",
				},
				ObjectMeta: types.ObjectMeta{
					Name:   volumeName,
					Labels: labels,
				},
				Spec: types.VolumeSpec{
					Size:      size,
					ServerID:  server,
					Labels:    labels,
					Automount: automount,
					Format:    format,
				},
			}
			return createVolume(volume)
		},
	}

	cmd.Flags().IntVar(&size, "size", config.DefaultVolumeSize, "Volume size in GB")
	cmd.Flags().StringVar(&server, "server", "", "Server to attach the volume to")
	cmd.Flags().StringToStringVar(&labels, "labels", map[string]string{}, "Volume labels")
	cmd.Flags().BoolVar(&automount, "automount", false, "Automount the volume")
	cmd.Flags().StringVar(&format, "format", config.DefaultVolumeFormat, "Volume format")
	if err := cmd.MarkFlagRequired("server"); err != nil {
		panic(err)
	}

	return cmd
}

func createVolume(volume *types.Volume) error {
	fmt.Printf("Creating volume %s with size %d\n",
		volume.ObjectMeta.Name,
		volume.Spec.Size)
	if volume.Spec.ServerID != "" {
		fmt.Printf("  server: %s\n", volume.Spec.ServerID)
	}
	if volume.Spec.Automount {
		fmt.Printf("  automount: %t\n", volume.Spec.Automount)
	}
	if volume.Spec.Format != "" {
		fmt.Printf("  format: %s\n", volume.Spec.Format)
	}
	labels := volume.ObjectMeta.Labels
	ttl := volume.Spec.TTL
	if ttl == "" {
		ttl = config.DefaultTTL
	}
	labels["delete_after"] = timeutil.FormatDeleteAfter(timeutil.TtlToDeleteAfter(ttl))
	labels["owner"] = labelutil.SanitizeValue(cfg.Owner)

	_, err := providerSvc.CreateVolume(options.VolumeCreateOpts{
		Name:       volume.ObjectMeta.Name,
		Size:       volume.Spec.Size,
		ServerName: volume.Spec.ServerName,
		Labels:     labels,
		Automount:  volume.Spec.Automount,
		Format:     volume.Spec.Format,
	})
	return err
}
