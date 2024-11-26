package cmd

import (
	"fmt"

	"github.com/pavelanni/labshop/internal/config"
	"github.com/pavelanni/labshop/internal/types"
	"github.com/pavelanni/labshop/internal/util/labelutil"
	"github.com/pavelanni/labshop/internal/util/timeutil"
	"github.com/spf13/cobra"
)

func NewCreateServerCmd() *cobra.Command {
	var (
		sshKeyNames []string
		serverType  string
		image       string
		location    string
		ttl         string
		labels      map[string]string
	)

	cmd := &cobra.Command{
		Use:   "server [name]",
		Short: "Create a new server",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			serverName := args[0]
			_, err := providerSvc.CreateServer(serverName, serverType, image, location, sshKeyNames, labels)
			return err
		},
	}

	cmd.Flags().StringSliceVar(&sshKeyNames, "ssh-keys", []string{}, "SSH key names to use (required)")
	cmd.Flags().StringVar(&serverType, "type", "cx22", "Server type")
	cmd.Flags().StringVar(&image, "image", "ubuntu-24.04", "Server image")
	cmd.Flags().StringVar(&location, "location", "fsn1", "Server location")
	cmd.Flags().StringVar(&ttl, "ttl", config.DefaultTTL, "Server TTL")
	cmd.Flags().StringToStringVar(&labels, "labels", map[string]string{}, "Server labels")
	if err := cmd.MarkFlagRequired("ssh-keys"); err != nil {
		panic(err)
	}

	return cmd
}

func createServer(server *types.Resource) error {
	// Access fields using map syntax
	fmt.Printf("Creating server %s with type %s, image %s, location %s, ssh keys %v\n",
		server.Metadata["name"],
		server.Spec["serverType"],
		server.Spec["image"],
		server.Spec["location"],
		server.Spec["keys"])

	keyNames := []string{}
	if keysInterface, ok := server.Spec["keys"]; ok {
		if keysSlice, ok := keysInterface.([]any); ok {
			for _, key := range keysSlice {
				if keyMap, ok := key.(map[string]any); ok {
					if name, ok := keyMap["name"]; ok {
						keyNames = append(keyNames, name.(string))
					}
				}
			}
		}
	}

	ttl := config.DefaultTTL
	if ttlInterface, ok := server.Spec["ttl"]; ok {
		if ttlStr, ok := ttlInterface.(string); ok {
			ttl = ttlStr
		}
	}

	stringLabels := make(map[string]string)
	if labelsInterface, ok := server.Metadata["labels"]; ok {
		if labels, ok := labelsInterface.(map[string]interface{}); ok {
			for k, v := range labels {
				stringLabels[k] = fmt.Sprint(v)
			}
		}
	}
	stringLabels["delete_after"] = timeutil.FormatDeleteAfter(timeutil.TtlToDeleteAfter(ttl))
	stringLabels["owner"] = labelutil.SanitizeValue(cfg.Owner)

	// DEBUG
	fmt.Printf("Creating SSH keys %v\n", keyNames)
	for _, key := range keyNames {
		err := createKey(&types.Resource{
			Metadata: map[string]any{
				"name":   key,
				"labels": stringLabels,
			},
		})
		if err != nil {
			return err
		}
	}
	_, err := providerSvc.CreateServer(
		server.Metadata["name"].(string),
		server.Spec["serverType"].(string),
		server.Spec["image"].(string),
		server.Spec["location"].(string),
		keyNames,
		stringLabels,
	)
	return err
}
