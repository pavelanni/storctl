package storage

import "github.com/pavelanni/storctl/internal/types"

type Storage interface {
	Save(lab *types.Lab) error
	Get(name string) (*types.Lab, error)
	List() ([]*types.Lab, error)
	Delete(name string) error
	Close() error
}
