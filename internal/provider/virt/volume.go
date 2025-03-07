package virt

import (
	"github.com/pavelanni/storctl/internal/provider/options"
	"github.com/pavelanni/storctl/internal/types"
)

func (p *VirtProvider) CreateVolume(opts options.VolumeCreateOpts) (*types.Volume, error) {
	return nil, nil
}

func (p *VirtProvider) GetVolume(name string) (*types.Volume, error) {
	return nil, nil
}

func (p *VirtProvider) AllVolumes() ([]*types.Volume, error) {
	return nil, nil
}

func (p *VirtProvider) ListVolumes(opts options.VolumeListOpts) ([]*types.Volume, error) {
	return nil, nil
}

func (p *VirtProvider) DeleteVolume(name string, force bool) *types.VolumeDeleteStatus {
	return &types.VolumeDeleteStatus{
		Deleted: false,
	}
}
