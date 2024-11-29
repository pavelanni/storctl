package cmd

import (
	"fmt"

	"github.com/pavelanni/labshop/internal/config"
	"github.com/pavelanni/labshop/internal/logger"
	"github.com/pavelanni/labshop/internal/provider/options"
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
					Type:        serverType,
					Image:       image,
					Location:    location,
					SSHKeyNames: sshKeyNames,
				},
			}
			result, err := createServer(server)
			if err != nil {
				return err
			}
			logger.Info("Server created successfully", "server", result.ObjectMeta.Name)
			return nil
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

func createServer(server *types.Server) (*types.Server, error) {
	// Access fields using map syntax
	fmt.Printf("Creating server %s with type %s, image %s, location %s, ssh keys %v\n",
		server.ObjectMeta.Name,
		server.Spec.Type,
		server.Spec.Image,
		server.Spec.Location,
		server.Spec.SSHKeyNames)

	ttl := config.DefaultTTL
	if server.Spec.TTL != "" {
		ttl = server.Spec.TTL
	}

	labels := server.ObjectMeta.Labels
	labels["delete_after"] = timeutil.FormatDeleteAfter(timeutil.TtlToDeleteAfter(ttl))
	labels["owner"] = labelutil.SanitizeValue(cfg.Owner)

	sshKeys := make([]*types.SSHKey, 0)
	for _, sshKeyName := range server.Spec.SSHKeyNames {
		keyExists, err := providerSvc.KeyExists(sshKeyName)
		if err != nil {
			return nil, err
		}
		if !keyExists { // Key not found, create it
			logger.Info("SSH key not found, creating",
				"key", sshKeyName)
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
	logger.Info("cloud-init user data", "data", cloudInitUserData)
	result, err := providerSvc.CreateServer(options.ServerCreateOpts{
		Name:     server.ObjectMeta.Name,
		Type:     server.Spec.Type,
		Image:    server.Spec.Image,
		Location: server.Spec.Location,
		SSHKeys:  sshKeys,
		Labels:   labels,
		UserData: cloudInitUserData,
	})

	return result, err
}
