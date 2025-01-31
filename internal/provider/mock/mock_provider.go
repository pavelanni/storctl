package mock

import (
	"github.com/pavelanni/storctl/internal/provider"
	"github.com/pavelanni/storctl/internal/provider/options"
	"github.com/pavelanni/storctl/internal/types"
)

// MockProvider implements the CloudProvider interface for testing
type MockProvider struct {
	NameFunc func() string
	// Function fields to customize behavior
	CreateServerFunc       func(opts options.ServerCreateOpts) (*types.Server, error)
	GetServerFunc          func(name string) (*types.Server, error)
	ListServersFunc        func(opts options.ServerListOpts) ([]*types.Server, error)
	AllServersFunc         func() ([]*types.Server, error)
	DeleteServerFunc       func(name string, force bool) *types.ServerDeleteStatus
	ServerToCreateOptsFunc func(server *types.Server) (options.ServerCreateOpts, error)
	CreateVolumeFunc       func(opts options.VolumeCreateOpts) (*types.Volume, error)
	GetVolumeFunc          func(name string) (*types.Volume, error)
	ListVolumesFunc        func(opts options.VolumeListOpts) ([]*types.Volume, error)
	AllVolumesFunc         func() ([]*types.Volume, error)
	DeleteVolumeFunc       func(name string, force bool) *types.VolumeDeleteStatus
	CreateLabOnCloudFunc   func(lab *types.Lab) error
	GetLabFromCloudFunc    func(name string) (*types.Lab, error)
	ListLabsFunc           func(opts options.LabListOpts) ([]*types.Lab, error)
	DeleteLabFromCloudFunc func(name string, force bool) *types.LabDeleteStatus
	SyncLabsFunc           func() error
	AllSSHKeysFunc         func() ([]*types.SSHKey, error)
	CreateSSHKeyFunc       func(opts options.SSHKeyCreateOpts) (*types.SSHKey, error)
	DeleteSSHKeyFunc       func(name string, force bool) *types.SSHKeyDeleteStatus
	GetSSHKeyFunc          func(name string) (*types.SSHKey, error)
	CloudKeyExistsFunc     func(name string) (bool, error)
	ListSSHKeysFunc        func(opts options.SSHKeyListOpts) ([]*types.SSHKey, error)
	KeyNamesToSSHKeysFunc  func(keyNames []string, opts options.SSHKeyCreateOpts) ([]*types.SSHKey, error)
}

// Ensure MockProvider implements CloudProvider interface
var _ provider.CloudProvider = &MockProvider{}

func (m *MockProvider) Name() string {
	if m.NameFunc != nil {
		return m.NameFunc()
	}
	return "mock"
}

// Implementation of interface methods
func (m *MockProvider) CreateServer(opts options.ServerCreateOpts) (*types.Server, error) {
	if m.CreateServerFunc != nil {
		return m.CreateServerFunc(opts)
	}
	return nil, nil
}

func (m *MockProvider) GetServer(name string) (*types.Server, error) {
	if m.GetServerFunc != nil {
		return m.GetServerFunc(name)
	}
	return nil, nil
}

func (m *MockProvider) ListServers(opts options.ServerListOpts) ([]*types.Server, error) {
	if m.ListServersFunc != nil {
		return m.ListServersFunc(opts)
	}
	return nil, nil
}

func (m *MockProvider) AllServers() ([]*types.Server, error) {
	if m.AllServersFunc != nil {
		return m.AllServersFunc()
	}
	return nil, nil
}

func (m *MockProvider) DeleteServer(name string, force bool) *types.ServerDeleteStatus {
	if m.DeleteServerFunc != nil {
		return m.DeleteServerFunc(name, force)
	}
	return &types.ServerDeleteStatus{}
}

func (m *MockProvider) ServerToCreateOpts(server *types.Server) (options.ServerCreateOpts, error) {
	if m.ServerToCreateOptsFunc != nil {
		return m.ServerToCreateOptsFunc(server)
	}
	return options.ServerCreateOpts{}, nil
}

func (m *MockProvider) CreateVolume(opts options.VolumeCreateOpts) (*types.Volume, error) {
	if m.CreateVolumeFunc != nil {
		return m.CreateVolumeFunc(opts)
	}
	return nil, nil
}

func (m *MockProvider) GetVolume(name string) (*types.Volume, error) {
	if m.GetVolumeFunc != nil {
		return m.GetVolumeFunc(name)
	}
	return nil, nil
}

func (m *MockProvider) ListVolumes(opts options.VolumeListOpts) ([]*types.Volume, error) {
	if m.ListVolumesFunc != nil {
		return m.ListVolumesFunc(opts)
	}
	return nil, nil
}

func (m *MockProvider) AllVolumes() ([]*types.Volume, error) {
	if m.AllVolumesFunc != nil {
		return m.AllVolumesFunc()
	}
	return nil, nil
}

func (m *MockProvider) DeleteVolume(name string, force bool) *types.VolumeDeleteStatus {
	if m.DeleteVolumeFunc != nil {
		return m.DeleteVolumeFunc(name, force)
	}
	return &types.VolumeDeleteStatus{}
}

func (m *MockProvider) CreateLabOnCloud(lab *types.Lab) error {
	if m.CreateLabOnCloudFunc != nil {
		return m.CreateLabOnCloudFunc(lab)
	}
	return nil
}

func (m *MockProvider) GetLabFromCloud(name string) (*types.Lab, error) {
	if m.GetLabFromCloudFunc != nil {
		return m.GetLabFromCloudFunc(name)
	}
	return nil, nil
}

func (m *MockProvider) ListLabs(opts options.LabListOpts) ([]*types.Lab, error) {
	if m.ListLabsFunc != nil {
		return m.ListLabsFunc(opts)
	}
	return nil, nil
}

func (m *MockProvider) DeleteLabFromCloud(name string, force bool) *types.LabDeleteStatus {
	if m.DeleteLabFromCloudFunc != nil {
		return m.DeleteLabFromCloudFunc(name, force)
	}
	return &types.LabDeleteStatus{}
}

func (m *MockProvider) SyncLabs() error {
	if m.SyncLabsFunc != nil {
		return m.SyncLabsFunc()
	}
	return nil
}

func (m *MockProvider) AllSSHKeys() ([]*types.SSHKey, error) {
	if m.AllSSHKeysFunc != nil {
		return m.AllSSHKeysFunc()
	}
	return nil, nil
}

func (m *MockProvider) CreateSSHKey(opts options.SSHKeyCreateOpts) (*types.SSHKey, error) {
	if m.CreateSSHKeyFunc != nil {
		return m.CreateSSHKeyFunc(opts)
	}
	return nil, nil
}

func (m *MockProvider) DeleteSSHKey(name string, force bool) *types.SSHKeyDeleteStatus {
	if m.DeleteSSHKeyFunc != nil {
		return m.DeleteSSHKeyFunc(name, force)
	}
	return &types.SSHKeyDeleteStatus{}
}

func (m *MockProvider) GetSSHKey(name string) (*types.SSHKey, error) {
	if m.GetSSHKeyFunc != nil {
		return m.GetSSHKeyFunc(name)
	}
	return nil, nil
}

func (m *MockProvider) CloudKeyExists(name string) (bool, error) {
	if m.CloudKeyExistsFunc != nil {
		return m.CloudKeyExistsFunc(name)
	}
	return false, nil
}

func (m *MockProvider) ListSSHKeys(opts options.SSHKeyListOpts) ([]*types.SSHKey, error) {
	if m.ListSSHKeysFunc != nil {
		return m.ListSSHKeysFunc(opts)
	}
	return nil, nil
}

func (m *MockProvider) KeyNamesToSSHKeys(keyNames []string, opts options.SSHKeyCreateOpts) ([]*types.SSHKey, error) {
	if m.KeyNamesToSSHKeysFunc != nil {
		return m.KeyNamesToSSHKeysFunc(keyNames, opts)
	}
	return nil, nil
}
