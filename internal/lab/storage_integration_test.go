package lab

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/pavelanni/storctl/internal/config"
	"github.com/pavelanni/storctl/internal/provider/mock"
	"github.com/pavelanni/storctl/internal/types"
)

// TestStorageBackendSwitching verifies that lab manager correctly uses configured storage backend
func TestStorageBackendSwitching(t *testing.T) {
	tests := []struct {
		name        string
		storageType string
		setupConfig func(t *testing.T) *config.Config
	}{
		{
			name:        "BoltDB backend",
			storageType: "local",
			setupConfig: func(t *testing.T) *config.Config {
				tempDir := t.TempDir()
				return &config.Config{
					Owner:        "test-user",
					Organization: "test-org",
					Email:        "test@example.com",
					Storage: config.StorageConfig{
						Type: "local",
						Local: config.LocalConfig{
							Path:   filepath.Join(tempDir, "test.db"),
							Bucket: "labs",
						},
					},
				}
			},
		},
		{
			name:        "PostgreSQL backend",
			storageType: "postgres",
			setupConfig: func(t *testing.T) *config.Config {
				return &config.Config{
					Owner:        "test-user",
					Organization: "test-org",
					Email:        "test@example.com",
					Storage: config.StorageConfig{
						Type: "postgres",
						Postgres: config.PostgresConfig{
							Host:     "localhost",
							Port:     "5432",
							Database: "storctl_dev",
							User:     "postgres",
							Password: "storctl",
						},
					},
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			cfg := tt.setupConfig(t)
			mockProvider := &mock.MockProvider{}

			// Create lab manager
			manager, err := NewManager(mockProvider, cfg)
			if err != nil {
				// PostgreSQL might not be running - skip this test
				if tt.storageType == "postgres" {
					t.Skipf("PostgreSQL not available: %v", err)
					return
				}
				t.Fatalf("failed to create manager: %v", err)
			}
			defer manager.Storage.Close()

			// Create a test lab (without actually creating VMs)
			// Use timestamp to avoid conflicts with soft-deleted labs in PostgreSQL
			labName := fmt.Sprintf("test-storage-%s-%d", tt.storageType, time.Now().Unix())
			lab := &types.Lab{
				TypeMeta: types.TypeMeta{
					APIVersion: "v1",
					Kind:       "Lab",
				},
				ObjectMeta: types.ObjectMeta{
					Name: labName,
					Labels: map[string]string{
						"test": "storage-integration",
					},
				},
				Spec: types.LabSpec{
					Provider: "mock",
					Location: "test",
					TTL:      "1h",
				},
				Status: types.LabStatus{
					State:   "testing",
					Owner:   "test-user",
					Created: time.Now(),
				},
			}

			// Test Save
			err = manager.Storage.Save(lab)
			if err != nil {
				t.Fatalf("failed to save lab with %s: %v", tt.storageType, err)
			}
			t.Logf("✓ Lab saved to %s backend", tt.storageType)

			// Test Get
			retrieved, err := manager.Storage.Get(lab.Name)
			if err != nil {
				t.Fatalf("failed to get lab from %s: %v", tt.storageType, err)
			}
			if retrieved.Name != lab.Name {
				t.Errorf("expected name %s, got %s", lab.Name, retrieved.Name)
			}
			t.Logf("✓ Lab retrieved from %s backend", tt.storageType)

			// Test List
			labs, err := manager.Storage.List()
			if err != nil {
				t.Fatalf("failed to list labs from %s: %v", tt.storageType, err)
			}
			found := false
			for _, l := range labs {
				if l.Name == lab.Name {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("lab %s not found in list from %s backend", lab.Name, tt.storageType)
			}
			t.Logf("✓ Lab listed from %s backend (%d total labs)", tt.storageType, len(labs))

			// Test Delete
			err = manager.Storage.Delete(lab.Name)
			if err != nil {
				t.Fatalf("failed to delete lab from %s: %v", tt.storageType, err)
			}
			t.Logf("✓ Lab deleted from %s backend", tt.storageType)

			// Verify deletion
			_, err = manager.Storage.Get(lab.Name)
			if err == nil {
				t.Errorf("expected error when getting deleted lab from %s, got nil", tt.storageType)
			}
			t.Logf("✓ Deletion verified on %s backend", tt.storageType)
		})
	}
}

// TestStorageBackendDefault verifies that manager defaults to BoltDB when type is empty
func TestStorageBackendDefault(t *testing.T) {
	tempDir := t.TempDir()
	cfg := &config.Config{
		Owner:        "test-user",
		Organization: "test-org",
		Email:        "test@example.com",
		Storage: config.StorageConfig{
			Type: "", // Empty - should default to local
			Local: config.LocalConfig{
				Path:   filepath.Join(tempDir, "test.db"),
				Bucket: "labs",
			},
		},
	}

	mockProvider := &mock.MockProvider{}
	manager, err := NewManager(mockProvider, cfg)
	if err != nil {
		t.Fatalf("failed to create manager with default storage: %v", err)
	}
	defer manager.Storage.Close()

	// Verify we can use it
	lab := &types.Lab{
		ObjectMeta: types.ObjectMeta{Name: "test-default"},
		Spec:       types.LabSpec{Provider: "mock"},
		Status:     types.LabStatus{Created: time.Now()},
	}

	if err := manager.Storage.Save(lab); err != nil {
		t.Fatalf("failed to save lab with default storage: %v", err)
	}

	if _, err := manager.Storage.Get("test-default"); err != nil {
		t.Fatalf("failed to get lab from default storage: %v", err)
	}

	t.Logf("✓ Default storage backend (BoltDB) works correctly")
}

// TestStorageCloseProperly verifies that Close() is called on storage
func TestStorageCloseProperly(t *testing.T) {
	tempDir := t.TempDir()
	cfg := &config.Config{
		Owner:        "test-user",
		Organization: "test-org",
		Email:        "test@example.com",
		Storage: config.StorageConfig{
			Type: "local",
			Local: config.LocalConfig{
				Path:   filepath.Join(tempDir, "test.db"),
				Bucket: "labs",
			},
		},
	}

	mockProvider := &mock.MockProvider{}
	manager, err := NewManager(mockProvider, cfg)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	// Save a lab
	lab := &types.Lab{
		ObjectMeta: types.ObjectMeta{Name: "test-close"},
		Spec:       types.LabSpec{Provider: "mock"},
		Status:     types.LabStatus{Created: time.Now()},
	}
	manager.Storage.Save(lab)

	// Close storage
	if err := manager.Storage.Close(); err != nil {
		t.Fatalf("failed to close storage: %v", err)
	}

	// Verify database file exists and is closed
	dbPath := cfg.Storage.Local.Path
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Errorf("database file does not exist after close: %s", dbPath)
	}

	t.Logf("✓ Storage closed properly, database file persisted")
}
