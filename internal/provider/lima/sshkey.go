package lima

import (
	"os"
	"path/filepath"

	"github.com/pavelanni/storctl/internal/provider/options"
	"github.com/pavelanni/storctl/internal/types"
)

const limaDir = ".lima"

// Lima creates a default key in ~/.lima/_config/user.pub
// This is the only key that is created by default and is used for all VMs

type LimaSSHKey struct {
	Name      string
	PublicKey string
	Labels    map[string]string
}

func (p *LimaProvider) CreateSSHKey(opts options.SSHKeyCreateOpts) (*types.SSHKey, error) {
	key, err := p.GetSSHKey("default")
	if err != nil {
		return nil, err
	}
	return key, nil
}

func (p *LimaProvider) GetSSHKey(name string) (*types.SSHKey, error) {
	// This always returns the default key located in ~/.lima/_config/user.pub
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	userKeyPath := filepath.Join(homeDir, limaDir, "_config", "user.pub")
	userKey, err := os.ReadFile(userKeyPath)
	if err != nil {
		return nil, err
	}
	return mapSSHKey(&LimaSSHKey{Name: "default", PublicKey: string(userKey)}), nil
}

func (p *LimaProvider) ListSSHKeys(opts options.SSHKeyListOpts) ([]*types.SSHKey, error) {
	defaultKey, err := p.GetSSHKey("default")
	if err != nil {
		return nil, err
	}
	keys := []*types.SSHKey{
		defaultKey,
	}
	return keys, nil
}

func (p *LimaProvider) AllSSHKeys() ([]*types.SSHKey, error) {
	defaultKey, err := p.GetSSHKey("default")
	if err != nil {
		return nil, err
	}
	keys := []*types.SSHKey{
		defaultKey,
	}
	return keys, nil
}

func (p *LimaProvider) DeleteSSHKey(name string, force bool) *types.SSHKeyDeleteStatus {
	return &types.SSHKeyDeleteStatus{
		Deleted: false,
	}
}

func (p *LimaProvider) CloudKeyExists(name string) (bool, error) {
	// This always returns true because the default key is always created
	return true, nil
}

func (p *LimaProvider) KeyNamesToSSHKeys(keyNames []string, opts options.SSHKeyCreateOpts) ([]*types.SSHKey, error) {
	// This always returns the default key because the default key is always created
	defaultKey, err := p.GetSSHKey("default")
	if err != nil {
		return nil, err
	}
	return []*types.SSHKey{defaultKey}, nil
}

func mapSSHKey(sk *LimaSSHKey) *types.SSHKey {
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
	}
}
