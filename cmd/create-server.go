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
			server := &types.Server{
				TypeMeta: types.TypeMeta{
					Kind:       "Server",
					APIVersion: "v1",
				},
				ObjectMeta: types.ObjectMeta{
					Name:   serverName,
					Labels: labels,
				},
				Spec: types.ServerSpec{
					ServerType:  serverType,
					Image:       image,
					Location:    location,
					Provider:    cfg.Provider.Name,
					SSHKeyNames: sshKeyNames,
				},
			}
			result, err := createServer(server)
			if err != nil {
				return err
			}
			fmt.Printf("Server created successfully: %s\n", result.ObjectMeta.Name)
			return nil
		},
	}

	cmd.Flags().StringSliceVar(&sshKeyNames, "ssh-keys", []string{}, "SSH key names to use (required)")
	cmd.Flags().StringVar(&serverType, "type", config.DefaultServerType, "Server type")
	cmd.Flags().StringVar(&image, "image", config.DefaultImage, "Server image")
	cmd.Flags().StringVar(&location, "location", config.DefaultLocation, "Server location")
	cmd.Flags().StringVar(&ttl, "ttl", config.DefaultTTL, "Server TTL")
	cmd.Flags().StringToStringVar(&labels, "labels", map[string]string{}, "Server labels")
	if err := cmd.MarkFlagRequired("ssh-keys"); err != nil {
		panic(err)
	}

	return cmd
}

func createServer(server *types.Server) (*types.Server, error) {
	// Access fields using map syntax
	fmt.Printf("Creating server %s with type %s, image %s, location %s, ssh keys %v\n",
		server.ObjectMeta.Name,
		server.Spec.ServerType,
		server.Spec.Image,
		server.Spec.Location,
		server.Spec.SSHKeyNames)

	if len(server.Spec.SSHKeyNames) == 0 {
		serverKeyName := server.ObjectMeta.Name + "-admin"
		fmt.Printf("No SSH keys provided, using default: %s\n", serverKeyName)
		server.Spec.SSHKeyNames = []string{serverKeyName}
	}
	ttl := config.DefaultTTL
	if server.Spec.TTL != "" {
		ttl = server.Spec.TTL
	}

	labels := server.ObjectMeta.Labels
	duration, err := timeutil.TtlToDuration(ttl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ttl: %w", err)
	}
	labels["delete_after"] = timeutil.FormatDeleteAfter(time.Now().Add(duration))
	labels["owner"] = labelutil.SanitizeValue(cfg.Owner)

	sshKeys := make([]*types.SSHKey, 0)
	for _, sshKeyName := range server.Spec.SSHKeyNames {
		keyExists, err := providerSvc.KeyExists(sshKeyName)
		if err != nil {
			return nil, err
		}
		if !keyExists {
			fmt.Printf("Creating new SSH key: %s\n", sshKeyName)
			newKey, err := createKey(&types.SSHKey{
				TypeMeta: types.TypeMeta{
					Kind: "SSHKey",
				},
				ObjectMeta: types.ObjectMeta{
					Name:   sshKeyName,
					Labels: labels,
				},
			})
			if err != nil {
				return nil, err
			}
			sshKeys = append(sshKeys, newKey)
		} else {
			providerKey, err := providerSvc.GetSSHKey(sshKeyName)
			if err != nil {
				return nil, err
			}
			sshKeys = append(sshKeys, providerKey)
		}
	}

	cloudInitUserData := fmt.Sprintf(config.DefaultCloudInitUserData, sshKeys[0].Spec.PublicKey)
	result, err := providerSvc.CreateServer(options.ServerCreateOpts{
		Name:     server.ObjectMeta.Name,
		Type:     server.Spec.ServerType,
		Image:    server.Spec.Image,
		Location: server.Spec.Location,
		Provider: server.Spec.Provider,
		SSHKeys:  sshKeys,
		Labels:   labels,
		UserData: cloudInitUserData,
	})

	return result, err
}
