package local

import (
	"encoding/json"
	"fmt"

	"log/slog"

	"github.com/pavelanni/storctl/internal/config"
	"github.com/pavelanni/storctl/internal/logger"
	"github.com/pavelanni/storctl/internal/types"
	"go.etcd.io/bbolt"
)

// Storage implements the Storage interface for boltdb
type Storage struct {
	db        *bbolt.DB
	labBucket []byte
	logger    *slog.Logger
}

func NewBboltDB(path string) (*bbolt.DB, error) {
	db, err := bbolt.Open(path, 0600, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open bbolt db: %w", err)
	}
	return db, nil
}

func New(cfg *config.Config) (*Storage, error) {
	db, err := NewBboltDB(cfg.Storage.Local.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to open bbolt from file %s: %w", cfg.Storage.Local.Path, err)
	}

	// Create bucket if it doesn't exist
	err = db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(cfg.Storage.Local.Bucket))
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create bucket %s: %w", cfg.Storage.Local.Bucket, err)
	}

	return &Storage{
		db:        db,
		labBucket: []byte(cfg.Storage.Local.Bucket),
		logger:    logger.Get(),
	}, nil
}

func (s *Storage) Save(lab *types.Lab) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(s.labBucket)
		data, err := json.Marshal(lab)
		if err != nil {
			return fmt.Errorf("failed to marshal lab: %w", err)
		}
		return b.Put([]byte(lab.Name), data)
	})
}

func (s *Storage) Get(name string) (*types.Lab, error) {
	var lab *types.Lab

	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(s.labBucket)
		data := b.Get([]byte(name))
		if data == nil {
			return fmt.Errorf("lab %s not found", name)
		}

		lab = &types.Lab{}
		if err := json.Unmarshal(data, lab); err != nil {
			return fmt.Errorf("failed to unmarshal lab: %w", err)
		}
		return nil
	})

	return lab, err
}

func (s *Storage) List(showDeleted bool) ([]*types.Lab, error) {
	// Note: BoltDB doesn't support soft deletes, so showDeleted parameter is ignored
	var labs []*types.Lab

	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(s.labBucket)
		if b == nil {
			return fmt.Errorf("labs bucket not found in database")
		}

		return b.ForEach(func(k, v []byte) error {
			lab := &types.Lab{}
			lab.Name = string(k)
			if err := json.Unmarshal(v, lab); err != nil {
				return fmt.Errorf("failed to unmarshal lab: %w", err)
			}
			labs = append(labs, lab)
			return nil
		})
	})

	return labs, err
}

func (s *Storage) Delete(name string) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		return tx.Bucket(s.labBucket).Delete([]byte(name))
	})
}

func (s *Storage) Close() error {
	return s.db.Close()
}
