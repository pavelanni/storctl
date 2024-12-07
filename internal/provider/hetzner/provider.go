package hetzner

import (
	"fmt"
	"log/slog"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/pavelanni/labshop/internal/config"
	"github.com/pavelanni/labshop/internal/logger"
	"go.etcd.io/bbolt"
)

type HetznerProvider struct {
	Client    *hcloud.Client
	config    *config.Config
	logger    *slog.Logger
	db        *bbolt.DB
	labBucket []byte
}

func New(cfg *config.Config) (*HetznerProvider, error) {
	token := cfg.Provider.Token
	if token == "" {
		return nil, fmt.Errorf("Hetzner API token is required")
	}

	// Create a new logger with the configured log level
	logLevel := logger.ParseLevel(cfg.LogLevel)
	providerLogger := logger.NewLogger(logLevel)

	client := hcloud.NewClient(hcloud.WithToken(token))
	p := &HetznerProvider{
		Client:    client,
		config:    cfg,
		logger:    providerLogger,
		labBucket: []byte("labs"),
	}
	// Open the database
	db, err := bbolt.Open("labs.db", 0600, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}
	p.db = db

	// Create the bucket
	err = db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(p.labBucket)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create bucket: %w", err)
	}

	return p, nil
}
