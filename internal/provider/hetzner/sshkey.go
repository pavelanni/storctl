package hetzner

import (
	"context"
	"fmt"
	"time"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/pavelanni/storctl/internal/provider/options"
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
		return nil, err
	}
	return mapSSHKey(sshKey), nil
}

func (p *HetznerProvider) GetSSHKey(name string) (*types.SSHKey, error) {
	p.logger.Debug("getting SSH key",
		"key", name)
	sshKey, _, err := p.Client.SSHKey.GetByName(context.Background(), name)
	if err != nil {
		return nil, err
	}
	if sshKey == nil {
		p.logger.Debug("SSH key not found",
			"key", name)
		return nil, fmt.Errorf("SSH key not found")
	}
	return mapSSHKey(sshKey), nil
}

func (p *HetznerProvider) ListSSHKeys(opts options.SSHKeyListOpts) ([]*types.SSHKey, error) {
	sshKeys, _, err := p.Client.SSHKey.List(context.Background(), hcloud.SSHKeyListOpts{
		ListOpts: hcloud.ListOpts{
			LabelSelector: opts.LabelSelector,
		},
	})
	if err != nil {
		return nil, err
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

	keyExists, err := p.KeyExists(name)
	if err != nil {
		return &types.SSHKeyDeleteStatus{
			Error: err,
		}
	}
	if !keyExists {
		p.logger.Debug("SSH key not found, skipping",
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
			Error: err,
		}
	}

	if !force {
		if deleteAfterStr, ok := sshKey.Labels["delete_after"]; ok {
			deleteAfter, err := time.Parse(time.RFC3339, deleteAfterStr)
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

	p.logger.Debug("deleting SSH key",
		"key", name)

	_, err = p.Client.SSHKey.Delete(context.Background(), sshKey)
	if err != nil {
		p.logger.Error("failed to delete SSH key",
			"key", name)
	}
	return &types.SSHKeyDeleteStatus{
		Deleted: true,
	}
}

func (p *HetznerProvider) KeyExists(name string) (bool, error) {
	sshKey, _, err := p.Client.SSHKey.GetByName(context.Background(), name)
	if err != nil {
		return false, fmt.Errorf("failed to check SSH key existence: %w", err)
	}
	return sshKey != nil, nil
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
