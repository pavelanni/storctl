package provider

import (
	"github.com/pavelanni/labshop/internal/provider/options"
	"github.com/pavelanni/labshop/internal/types"
)

type CloudProvider interface {
	// Server operations
	CreateServer(opts options.ServerCreateOpts) (*types.Server, error)
	GetServer(name string) (*types.Server, error)
	ListServers(opts options.ServerListOpts) ([]*types.Server, error)
	AllServers() ([]*types.Server, error)
	DeleteServer(name string, force bool) error

	// Volume operations
	CreateVolume(opts options.VolumeCreateOpts) (*types.Volume, error)
	GetVolume(name string) (*types.Volume, error)
	ListVolumes(opts options.VolumeListOpts) ([]*types.Volume, error)
	AllVolumes() ([]*types.Volume, error)
	DeleteVolume(name string, force bool) error

	// Lab operations
	CreateLab(name string, template string) error
	GetLab(name string) (*types.Lab, error)
	ListLabs(opts options.LabListOpts) ([]*types.Lab, error)
	DeleteLab(name string, force bool) error

	// SSH Key operations
	CreateSSHKey(opts options.SSHKeyCreateOpts) (*types.SSHKey, error)
	GetSSHKey(name string) (*types.SSHKey, error)
	ListSSHKeys(opts options.SSHKeyListOpts) ([]*types.SSHKey, error)
	AllSSHKeys() ([]*types.SSHKey, error)
	DeleteSSHKey(name string, force bool) error
	KeyExists(name string) (bool, error)
}
