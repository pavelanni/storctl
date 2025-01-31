package provider

import (
	"github.com/pavelanni/storctl/internal/provider/options"
	"github.com/pavelanni/storctl/internal/types"
)

type CloudProvider interface {
	Name() string
	// Server operations
	CreateServer(opts options.ServerCreateOpts) (*types.Server, error)
	GetServer(name string) (*types.Server, error)
	ListServers(opts options.ServerListOpts) ([]*types.Server, error)
	AllServers() ([]*types.Server, error)
	DeleteServer(name string, force bool) *types.ServerDeleteStatus
	ServerToCreateOpts(server *types.Server) (options.ServerCreateOpts, error)
	// Volume operations
	CreateVolume(opts options.VolumeCreateOpts) (*types.Volume, error)
	GetVolume(name string) (*types.Volume, error)
	ListVolumes(opts options.VolumeListOpts) ([]*types.Volume, error)
	AllVolumes() ([]*types.Volume, error)
	DeleteVolume(name string, force bool) *types.VolumeDeleteStatus

	// SSH Key operations
	CreateSSHKey(opts options.SSHKeyCreateOpts) (*types.SSHKey, error)
	GetSSHKey(name string) (*types.SSHKey, error)
	ListSSHKeys(opts options.SSHKeyListOpts) ([]*types.SSHKey, error)
	AllSSHKeys() ([]*types.SSHKey, error)
	DeleteSSHKey(name string, force bool) *types.SSHKeyDeleteStatus
	CloudKeyExists(name string) (bool, error)
	KeyNamesToSSHKeys(keyNames []string, opts options.SSHKeyCreateOpts) ([]*types.SSHKey, error)
}
