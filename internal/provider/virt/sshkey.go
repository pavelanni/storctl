package virt

import (
	"github.com/pavelanni/storctl/internal/provider/options"
	"github.com/pavelanni/storctl/internal/types"
)

func (p *VirtProvider) CreateSSHKey(opts options.SSHKeyCreateOpts) (*types.SSHKey, error) {
	return nil, nil
}

func (p *VirtProvider) GetSSHKey(name string) (*types.SSHKey, error) {
	return nil, nil
}

func (p *VirtProvider) AllSSHKeys() ([]*types.SSHKey, error) {
	return nil, nil
}

func (p *VirtProvider) ListSSHKeys(opts options.SSHKeyListOpts) ([]*types.SSHKey, error) {
	return nil, nil
}

func (p *VirtProvider) DeleteSSHKey(name string, force bool) *types.SSHKeyDeleteStatus {
	return &types.SSHKeyDeleteStatus{
		Deleted: false,
	}
}

func (p *VirtProvider) CloudKeyExists(name string) (bool, error) {
	return false, nil
}

func (p *VirtProvider) KeyNamesToSSHKeys(keyNames []string, opts options.SSHKeyCreateOpts) ([]*types.SSHKey, error) {
	return nil, nil
}
