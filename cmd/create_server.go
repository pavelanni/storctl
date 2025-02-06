package cmd

import (
	"fmt"
	"time"

	"github.com/pavelanni/storctl/internal/config"
	"github.com/pavelanni/storctl/internal/provider/options"
	"github.com/pavelanni/storctl/internal/ssh"
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
		provider    string
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
					Provider:    provider,
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

	cmd.Flags().StringSliceVar(&sshKeyNames, "ssh-keys", []string{}, "SSH key names to use; if not provided, the admin key will be created")
	cmd.Flags().StringVar(&serverType, "type", config.DefaultServerType, "Server type")
	cmd.Flags().StringVar(&image, "image", config.DefaultImage, "Server image")
	cmd.Flags().StringVar(&provider, "provider", config.DefaultLocalProvider, "Server provider")
	cmd.Flags().StringVar(&location, "location", config.DefaultLocalLocation, "Server location")
	cmd.Flags().StringVar(&ttl, "ttl", config.DefaultTTL, "Server TTL")
	cmd.Flags().StringToStringVar(&labels, "labels", map[string]string{}, "Server labels")

	return cmd
}

func createServer(server *types.Server) (*types.Server, error) {
	err := initProvider(server.Spec.Provider)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize provider: %w", err)
	}

	var sshKeys []*types.SSHKey
	sshManager := ssh.NewManager(cfg)
	// no ssh keys provided, use the admin key
	if len(server.Spec.SSHKeyNames) == 0 {
		serverKeyName := server.ObjectMeta.Name + "-admin"
		fmt.Printf("No SSH keys provided, using default: %s\n", serverKeyName)
		server.Spec.SSHKeyNames = []string{serverKeyName}
	}
	// Access fields using map syntax
	fmt.Printf("Creating server %s with type %s, image %s, provider %s, location %s, ssh keys %v\n",
		server.ObjectMeta.Name,
		server.Spec.ServerType,
		server.Spec.Image,
		server.Spec.Provider,
		server.Spec.Location,
		server.Spec.SSHKeyNames)

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

	if server.Spec.Provider != "lima" {
		// create the ssh keys locally
		for _, sshKeyName := range server.Spec.SSHKeyNames {
			_, err := sshManager.CreateLocalKeyPair(sshKeyName)
			if err != nil {
				return nil, fmt.Errorf("failed to create local ssh key: %w", err)
			}
		}
		sshKeys, err = providerSvc.KeyNamesToSSHKeys(server.Spec.SSHKeyNames, options.SSHKeyCreateOpts{
			Labels: labels,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to upload ssh keys to the cloud: %w", err)
		}
	}

	opts, err := providerSvc.ServerToCreateOpts(server)
	if err != nil {
		return nil, fmt.Errorf("failed to convert server to create opts: %w", err)
	}
	if server.Spec.Provider != "lima" {
		opts.SSHKeys = sshKeys
	}
	result, err := providerSvc.CreateServer(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create server: %w", err)
	}

	return result, nil
}
