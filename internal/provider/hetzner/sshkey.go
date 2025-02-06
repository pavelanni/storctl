package hetzner

import (
	"context"
	"fmt"
	"time"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/pavelanni/storctl/internal/config"
	"github.com/pavelanni/storctl/internal/provider/options"
	"github.com/pavelanni/storctl/internal/ssh"
	"github.com/pavelanni/storctl/internal/types"
	"github.com/pavelanni/storctl/internal/util/timeutil"
)

func (p *HetznerProvider) CreateSSHKey(opts options.SSHKeyCreateOpts) (*types.SSHKey, error) {
	p.logger.Debug("creating SSH key",
		"name", opts.Name,
		"public_key", opts.PublicKey)
	sshKey, _, err := p.Client.SSHKey.Create(context.Background(), hcloud.SSHKeyCreateOpts{
		Name:      opts.Name,
		PublicKey: opts.PublicKey,
		Labels:    opts.Labels,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating SSH key: %w", err)
	}
	return mapSSHKey(sshKey), nil
}

func (p *HetznerProvider) GetSSHKey(name string) (*types.SSHKey, error) {
	p.logger.Debug("getting SSH key",
		"key", name)
	sshKey, _, err := p.Client.SSHKey.GetByName(context.Background(), name)
	if err != nil {
		return nil, fmt.Errorf("error getting SSH key: %w", err)
	}
	if sshKey == nil {
		p.logger.Debug("SSH key not found",
			"key", name)
		return nil, fmt.Errorf("SSH key not found")
	}
	p.logger.Debug("SSH key found",
		"key", name,
		"public_key", sshKey.PublicKey)
	return mapSSHKey(sshKey), nil
}

func (p *HetznerProvider) ListSSHKeys(opts options.SSHKeyListOpts) ([]*types.SSHKey, error) {
	sshKeys, _, err := p.Client.SSHKey.List(context.Background(), hcloud.SSHKeyListOpts{
		ListOpts: hcloud.ListOpts{
			LabelSelector: opts.LabelSelector,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("error listing SSH keys: %w", err)
	}
	return mapSSHKeys(sshKeys), nil
}

func (p *HetznerProvider) AllSSHKeys() ([]*types.SSHKey, error) {
	sshKeys, err := p.Client.SSHKey.All(context.Background())
	if err != nil {
		return nil, err
	}
	return mapSSHKeys(sshKeys), nil
}

func (p *HetznerProvider) DeleteSSHKey(name string, force bool) *types.SSHKeyDeleteStatus {
	if name == "" {
		return &types.SSHKeyDeleteStatus{
			Error: fmt.Errorf("empty SSH key name provided"),
		}
	}

	keyExists, err := p.CloudKeyExists(name)
	if err != nil {
		return &types.SSHKeyDeleteStatus{
			Error: fmt.Errorf("error checking if SSH key exists: %w", err),
		}
	}
	if !keyExists {
		p.logger.Debug("SSH key not found on the cloud, skipping",
			"key", name)
		return &types.SSHKeyDeleteStatus{
			Deleted: true,
		}
	}

	sshKey, _, err := p.Client.SSHKey.GetByName(context.Background(), name)
	if err != nil {
		p.logger.Error("failed to get SSH key",
			"key", name)
		return &types.SSHKeyDeleteStatus{
			Error: fmt.Errorf("error getting SSH key: %w", err),
		}
	}

	if !force {
		if deleteAfterStr, ok := sshKey.Labels["delete_after"]; ok {
			deleteAfter := timeutil.ParseDeleteAfter(deleteAfterStr)
			if err == nil && time.Now().UTC().Before(deleteAfter) {
				p.logger.Warn("key not ready for deletion",
					"key", name,
					"delete_after", deleteAfter.Format("2006-01-02 15:04:05"))
				return &types.SSHKeyDeleteStatus{
					DeleteAfter: deleteAfter,
				}
			}
		}
	}

	p.logger.Debug("deleting cloud SSH key",
		"key", name)

	_, err = p.Client.SSHKey.Delete(context.Background(), sshKey)
	if err != nil {
		p.logger.Error("failed to delete cloud SSH key",
			"key", name)
		return &types.SSHKeyDeleteStatus{
			Error: fmt.Errorf("error deleting SSH key: %w", err),
		}
	}
	return &types.SSHKeyDeleteStatus{
		Deleted: true,
	}
}

func (p *HetznerProvider) CloudKeyExists(name string) (bool, error) {
	// check if the cloud key exists
	cloudKey, _, err := p.Client.SSHKey.GetByName(context.Background(), name)
	if err != nil {
		return false, fmt.Errorf("failed to check SSH key existence: %w", err)
	}

	return cloudKey != nil, nil
}

// KeyNamesToSSHKeys converts a list of SSH key names to a list of SSH keys
// It will upload local SSH keys to the cloud if they don't exist
// It adds the default admin key to the list
func (p *HetznerProvider) KeyNamesToSSHKeys(keyNames []string, opts options.SSHKeyCreateOpts) ([]*types.SSHKey, error) {
	sshManager := ssh.NewManager(p.config)
	sshKeys := make([]*types.SSHKey, 0)
	adminKey, err := p.GetSSHKey(config.DefaultAdminKeyName)
	if err != nil {
		return nil, fmt.Errorf("failed to get admin key: %w", err)
	}
	sshKeys = append(sshKeys, adminKey)

	for _, keyName := range keyNames {
		cloudKeyExists, err := p.CloudKeyExists(keyName)
		if err != nil {
			return nil, fmt.Errorf("error checking if SSH key exists: %w", err)
		}
		if !cloudKeyExists {
			// check if the key exists locally
			localKeyExists, err := sshManager.LocalKeyExists(keyName)
			if err != nil {
				return nil, err
			}
			if !localKeyExists {
				fmt.Printf("SSH key %s not found locally, skipping it\n", keyName)
				continue
			}
			pubKey, err := sshManager.ReadLocalPublicKey(keyName)
			if err != nil {
				return nil, fmt.Errorf("failed to read local public key: %w", err)
			}
			opts.PublicKey = pubKey
			newKey, err := p.CreateSSHKey(opts)
			if err != nil {
				return nil, fmt.Errorf("failed to create SSH key: %w", err)
			}
			sshKeys = append(sshKeys, newKey)
		}
	}
	return sshKeys, nil
}

func mapSSHKey(sk *hcloud.SSHKey) *types.SSHKey {
	if sk == nil {
		return nil
	}

	return &types.SSHKey{
		TypeMeta: types.TypeMeta{
			APIVersion: "v1",
			Kind:       "SSHKey",
		},
		ObjectMeta: types.ObjectMeta{
			Name: sk.Name,
		},
		Spec: types.SSHKeySpec{
			PublicKey: sk.PublicKey,
			Labels:    sk.Labels,
		},
		Status: types.SSHKeyStatus{
			Created:     sk.Created,
			DeleteAfter: timeutil.ParseDeleteAfter(sk.Labels["delete_after"]),
		},
	}
}

func mapSSHKeys(sshKeys []*hcloud.SSHKey) []*types.SSHKey {
	if sshKeys == nil {
		return nil
	}

	result := make([]*types.SSHKey, len(sshKeys))
	for i, sk := range sshKeys {
		result[i] = mapSSHKey(sk)
	}
	return result
}
