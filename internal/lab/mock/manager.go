package mock

import (
	"github.com/pavelanni/storctl/internal/lab"
	"github.com/pavelanni/storctl/internal/types"
)

// Manager implements lab.ManagerSvc for testing
type Manager struct {
	*lab.ManagerSvc  // Embed the ManagerSvc
	ListFunc         func() ([]*types.Lab, error)
	GetFunc          func(name string) (*types.Lab, error)
	GetFromCloudFunc func(name string) (*types.Lab, error)
	CreateFunc       func(lab *types.Lab) error
	DeleteFunc       func(name string, force bool) error
}

func (m *Manager) List() ([]*types.Lab, error) {
	return m.ListFunc()
}

func (m *Manager) Get(name string) (*types.Lab, error) {
	return m.GetFunc(name)
}

func (m *Manager) Create(lab *types.Lab) error {
	return m.CreateFunc(lab)
}

func (m *Manager) Delete(name string, force bool) error {
	return m.DeleteFunc(name, force)
}

func (m *Manager) GetFromCloud(name string) (*types.Lab, error) {
	return m.GetFromCloudFunc(name)
}
