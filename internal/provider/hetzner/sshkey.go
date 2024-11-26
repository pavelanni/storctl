package hetzner

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/pavelanni/labshop/internal/types"
	"github.com/pavelanni/labshop/internal/util/timeutil"
)

func (p *HetznerProvider) CreateSSHKey(name string, publicKey string, labels map[string]string) (*types.SSHKey, error) {
	p.logger.Info("creating SSH key",
		"name", name,
		"public_key", publicKey)
	sshKey, _, err := p.Client.SSHKey.Create(context.Background(), hcloud.SSHKeyCreateOpts{
		Name:      name,
		PublicKey: publicKey,
		Labels:    labels,
	})
	if err != nil {
		return nil, err
	}
	return mapSSHKey(sshKey), nil
}

func (p *HetznerProvider) GetSSHKey(name string) (*types.SSHKey, error) {
	p.logger.Info("getting SSH key",
		"key", name)
	sshKey, _, err := p.Client.SSHKey.GetByName(context.Background(), name)
	if err != nil {
		return nil, err
	}
	if sshKey == nil {
		p.logger.Info("SSH key not found",
			"key", name)
		return nil, fmt.Errorf("SSH key not found")
	}
	return mapSSHKey(sshKey), nil
}

func (p *HetznerProvider) AllSSHKeys() ([]*types.SSHKey, error) {
	sshKeys, err := p.Client.SSHKey.All(context.Background())
	if err != nil {
		return nil, err
	}
	return mapSSHKeys(sshKeys), nil
}

func (p *HetznerProvider) DeleteSSHKey(name string, force bool) error {
	if name == "" {
		return fmt.Errorf("empty SSH key name provided")
	}

	sshKey, _, err := p.Client.SSHKey.GetByName(context.Background(), name)
	if err != nil {
		p.logger.Error("failed to get SSH key",
			"key", name)
		return err
	}

	if sshKey == nil {
		p.logger.Info("SSH key not found, skipping",
			"key", name)
		return nil
	}

	if !force {
		if deleteAfterStr, ok := sshKey.Labels["delete_after"]; ok {
			deleteAfter, err := time.Parse(time.RFC3339, deleteAfterStr)
			if err == nil && time.Now().Before(deleteAfter) {
				p.logger.Warn("key not ready for deletion",
					"key", name,
					"delete_after", deleteAfter.Format("2006-01-02 15:04:05"))
				return fmt.Errorf("key %s is not ready for deletion until %s",
					name, deleteAfter.Format("2006-01-02 15:04:05"))
			}
		}
	}

	p.logger.Info("deleting SSH key",
		"key", name)

	_, err = p.Client.SSHKey.Delete(context.Background(), sshKey)
	if err != nil {
		p.logger.Error("failed to delete SSH key",
			"key", name)
	}
	return err
}

func (p *HetznerProvider) KeyExists(name string) (bool, error) {
	sshKey, _, err := p.Client.SSHKey.GetByName(context.Background(), name)
	return sshKey != nil, err
}

func mapSSHKey(sk *hcloud.SSHKey) *types.SSHKey {
	if sk == nil {
		return nil
	}

	return &types.SSHKey{
		ID:          strconv.FormatInt(sk.ID, 10),
		Name:        sk.Name,
		Fingerprint: sk.Fingerprint,
		PublicKey:   sk.PublicKey,
		Labels:      sk.Labels,
		Created:     sk.Created,
		DeleteAfter: timeutil.ParseDeleteAfter(sk.Labels["delete_after"]),
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
