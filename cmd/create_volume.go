package cmd

import (
	"fmt"
	"time"

	"github.com/pavelanni/storctl/internal/config"
	"github.com/pavelanni/storctl/internal/provider/options"
	"github.com/pavelanni/storctl/internal/types"
	"github.com/pavelanni/storctl/internal/util/labelutil"
	"github.com/pavelanni/storctl/internal/util/timeutil"
	"github.com/spf13/cobra"
)

func NewCreateVolumeCmd() *cobra.Command {
	var (
		size      int
		server    string
		labels    map[string]string
		automount bool
		format    string
		provider  string
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
					Provider:  provider,
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
	cmd.Flags().StringVar(&provider, "provider", config.DefaultLocalProvider, "Provider")
	return cmd
}

func createVolume(volume *types.Volume) error {
	err := initProvider(volume.Spec.Provider)
	if err != nil {
		return fmt.Errorf("failed to initialize provider: %w", err)
	}
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
	duration, err := timeutil.TtlToDuration(ttl)
	if err != nil {
		return fmt.Errorf("failed to parse ttl: %w", err)
	}
	labels["delete_after"] = timeutil.FormatDeleteAfter(time.Now().Add(duration))
	labels["owner"] = labelutil.SanitizeValue(cfg.Owner)

	_, err = providerSvc.CreateVolume(options.VolumeCreateOpts{
		Name:       volume.ObjectMeta.Name,
		Size:       volume.Spec.Size,
		ServerName: volume.Spec.ServerName,
		Labels:     labels,
		Automount:  volume.Spec.Automount,
		Format:     volume.Spec.Format,
	})
	return err
}
