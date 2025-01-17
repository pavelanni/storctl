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

func NewCreateKeyCmd() *cobra.Command {
	var labels map[string]string
	var ttl string

	cmd := &cobra.Command{
		Use: "key [name]",

		Short: "Create and upload an SSH key pair",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			keyName := args[0]
			keyResource := &types.SSHKey{
				TypeMeta: types.TypeMeta{
					APIVersion: "v1",
					Kind:       "SSHKey",
				},
				ObjectMeta: types.ObjectMeta{
					Name:   keyName,
					Labels: labels,
				},
			}
			key, err := createKey(keyResource)
			if err != nil {
				return err
			}
			fmt.Printf("SSH key created successfully: %s\n", key.ObjectMeta.Name)
			return nil
		},
	}

	cmd.Flags().StringToStringVar(&labels, "labels", map[string]string{}, "SSH key labels")
	cmd.Flags().StringVar(&ttl, "ttl", config.DefaultTTL, "Time to live for the key")
	return cmd
}

func createKey(key *types.SSHKey) (*types.SSHKey, error) {
	keyManager := ssh.NewManager(cfg)
	keyName := key.ObjectMeta.Name
	if keyName == "" {
		return nil, fmt.Errorf("key name is required")
	}

	// If public key is not provided, generate a new key pair
	if key.Spec.PublicKey == "" {
		// Check if the key already exists locally
		// create the key pair if it doesn't exist
		localKeyExists, err := keyManager.LocalKeyExists(keyName)
		if err != nil {
			return nil, fmt.Errorf("failed to check if key exists locally: %w", err)
		}
		if !localKeyExists {
			fmt.Printf("Creating keypair %s locally\n", keyName)
			pubKey, err := keyManager.CreateLocalKeyPair(keyName)
			if err != nil {
				return nil, fmt.Errorf("failed to create keypair: %w", err)
			}
			key.Spec.PublicKey = pubKey
		} else {
			pubKey, err := keyManager.ReadLocalPublicKey(keyName)
			if err != nil {
				return nil, fmt.Errorf("failed to read local public key: %w", err)
			}
			key.Spec.PublicKey = pubKey
		}
	}

	// Check if key already exists on the provider
	keyExists, err := providerSvc.CloudKeyExists(keyName)
	if err != nil {
		return nil, fmt.Errorf("failed to check if key exists on provider: %w", err)
	}
	if keyExists {
		fmt.Printf("SSH key %s already exists on the provider\n", keyName)
		cloudKey, err := providerSvc.GetSSHKey(keyName)
		if err != nil {
			return nil, fmt.Errorf("failed to get key from provider: %w", err)
		}
		// is it the same key?
		if cloudKey.Spec.PublicKey == key.Spec.PublicKey {
			return key, nil
		} else {
			fmt.Printf("SSH key %s already exists on the provider but is different from the local key. Replacing it.\n", keyName)
			status := providerSvc.DeleteSSHKey(keyName, true)
			if status.Error != nil {
				return nil, fmt.Errorf("failed to delete key from provider: %w", status.Error)
			}
		}
	}

	fmt.Printf("Creating SSH key %s on provider\n", keyName)
	labels := key.ObjectMeta.Labels
	var ttl string
	if key.Spec.TTL == "" {
		ttl = config.DefaultTTL
	} else {
		ttl = key.Spec.TTL
	}
	duration, err := timeutil.TtlToDuration(ttl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ttl: %w", err)
	}
	labels["delete_after"] = timeutil.FormatDeleteAfter(time.Now().Add(duration))
	labels["owner"] = labelutil.SanitizeValue(cfg.Owner)
	// Upload public key to provider
	key, err = providerSvc.CreateSSHKey(options.SSHKeyCreateOpts{
		Name:      keyName,
		PublicKey: key.Spec.PublicKey,
		Labels:    labels,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload public key: %w", err)
	}

	fmt.Printf("SSH key uploaded to provider: %s\n", keyName)
	return key, nil
}
