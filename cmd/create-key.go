package cmd

import (
	"crypto"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pavelanni/storctl/internal/config"
	"github.com/pavelanni/storctl/internal/provider/options"
	"github.com/pavelanni/storctl/internal/types"
	"github.com/pavelanni/storctl/internal/util/labelutil"
	"github.com/pavelanni/storctl/internal/util/timeutil"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
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
	keyName := key.ObjectMeta.Name
	if keyName == "" {
		return nil, fmt.Errorf("key name is required")
	}

	// Check if key already exists locally
	keysDir := filepath.Join(os.Getenv("HOME"), config.DefaultConfigDir, config.KeysDir)
	localKeyPath := filepath.Join(keysDir, keyName)
	if _, err := os.Stat(localKeyPath); err == nil {
		return nil, fmt.Errorf("key %s already exists locally at %s", keyName, localKeyPath)
	}

	// Check if key already exists on the provider
	exists, err := providerSvc.KeyExists(keyName)
	if err != nil {
		return nil, fmt.Errorf("failed to check if key exists on provider: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("key %s already exists on the provider", keyName)
	}

	fmt.Printf("Creating key %s\n", keyName)
	labels := key.ObjectMeta.Labels
	var ttl string
	if key.Spec.TTL == "" {
		ttl = config.DefaultTTL
	} else {
		ttl = key.Spec.TTL
	}
	labels["delete_after"] = timeutil.FormatDeleteAfter(timeutil.TtlToDeleteAfter(ttl))
	labels["owner"] = labelutil.SanitizeValue(cfg.Owner)
	pubKeyString := key.Spec.PublicKey
	// If public key is not provided, generate a new key pair
	if pubKeyString == "" {
		// Generate the key pair
		pubKey, privKey, err := generateED25519KeyPair(keyName)
		if err != nil {
			return nil, fmt.Errorf("failed to generate key pair: %w", err)
		}

		// Save the keys locally
		keysDir := filepath.Join(os.Getenv("HOME"), config.DefaultConfigDir, config.KeysDir)
		if err := os.MkdirAll(keysDir, 0700); err != nil {
			return nil, fmt.Errorf("failed to create keys directory: %w", err)
		}

		// Save private key
		privKeyPath := filepath.Join(keysDir, keyName)
		if err := os.WriteFile(privKeyPath, privKey, 0600); err != nil {
			return nil, fmt.Errorf("failed to save private key: %w", err)
		}

		// Save public key
		pubKeyPath := filepath.Join(keysDir, keyName+".pub")
		if err := os.WriteFile(pubKeyPath, pubKey, 0644); err != nil {
			return nil, fmt.Errorf("failed to save public key: %w", err)
		}
		pubKeyString = string(pubKey)
		fmt.Printf("SSH key pair created successfully: %s\n", keyName)
		key.Spec.PublicKey = pubKeyString
	}

	// Upload public key to provider
	key, err = providerSvc.CreateSSHKey(options.SSHKeyCreateOpts{
		Name:      keyName,
		PublicKey: pubKeyString,
		Labels:    labels,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload public key: %w", err)
	}

	fmt.Printf("SSH key uploaded to provider: %s\n", keyName)
	return key, nil
}

// generateED25519KeyPair generates a new ED25519 keypair.
// Returns public key in OpenSSH format and private key in PEM format as byte slices.
func generateED25519KeyPair(comment string) (publicKey, privateKey []byte, err error) {
	// Generate the keypair
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate ED25519 keypair: %w", err)
	}

	// Convert to SSH private key format and encode as PEM
	pemBlock, err := ssh.MarshalPrivateKey(crypto.PrivateKey(priv), comment)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal private key: %w", err)
	}

	// Encode private key in PEM format
	privateKey = pem.EncodeToMemory(pemBlock)
	if privateKey == nil {
		return nil, nil, fmt.Errorf("failed to encode private key")
	}

	// Generate the public key
	sshPub, err := ssh.NewPublicKey(pub)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create public key: %w", err)
	}

	// Format public key in OpenSSH format: "ssh-ed25519 <base64> comment"
	pubKey := fmt.Sprintf("%s %s", sshPub.Type(),
		base64.StdEncoding.EncodeToString(sshPub.Marshal()))
	if comment != "" {
		pubKey = fmt.Sprintf("%s %s", pubKey, comment)
	}

	return []byte(pubKey), privateKey, nil
}
