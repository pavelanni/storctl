package postgres

import (
	"testing"
	"time"

	"github.com/pavelanni/storctl/internal/config"
	"github.com/pavelanni/storctl/internal/types"
)

// getTestConfig returns a config for testing against local PostgreSQL
func getTestConfig() *config.Config {
	return &config.Config{
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
}

// createTestLab creates a sample lab for testing
func createTestLab(name string) *types.Lab {
	return &types.Lab{
		TypeMeta: types.TypeMeta{
			APIVersion: "v1",
			Kind:       "Lab",
		},
		ObjectMeta: types.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"owner":   "test-user",
				"project": "test-project",
			},
		},
		Spec: types.LabSpec{
			Provider: "hetzner",
			Location: "hel1",
			TTL:      "1h",
		},
		Status: types.LabStatus{
			State:   "running",
			Owner:   "test-user",
			Created: time.Now(),
		},
	}
}

// cleanupTestData removes test data after tests
func cleanupTestData(t *testing.T, storage *Storage, labNames ...string) {
	for _, name := range labNames {
		if err := storage.Delete(name); err != nil {
			t.Logf("Warning: failed to cleanup lab %s: %v", name, err)
		}
	}
}

func TestSaveAndGet(t *testing.T) {
	// Setup
	cfg := getTestConfig()
	storage, err := New(cfg)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}
	defer storage.Close()

	labName := "test-save-get"
	lab := createTestLab(labName)
	defer cleanupTestData(t, storage, labName)

	// Test Save
	err = storage.Save(lab)
	if err != nil {
		t.Fatalf("failed to save lab: %v", err)
	}
	t.Logf("✓ Lab saved: %s", lab.Name)

	// Test Get
	retrieved, err := storage.Get(labName)
	if err != nil {
		t.Fatalf("failed to get lab: %v", err)
	}

	// Verify fields
	if retrieved.Name != lab.Name {
		t.Errorf("expected name %s, got %s", lab.Name, retrieved.Name)
	}
	if retrieved.ObjectMeta.Labels["owner"] != "test-user" {
		t.Errorf("expected owner test-user, got %s", retrieved.ObjectMeta.Labels["owner"])
	}
	if retrieved.Spec.Provider != "hetzner" {
		t.Errorf("expected provider hetzner, got %s", retrieved.Spec.Provider)
	}
	if retrieved.Status.State != "running" {
		t.Errorf("expected state running, got %s", retrieved.Status.State)
	}

	t.Logf("✓ Lab retrieved and verified: %s", retrieved.Name)
}

func TestList(t *testing.T) {
	// Setup
	cfg := getTestConfig()
	storage, err := New(cfg)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}
	defer storage.Close()

	// Create multiple test labs
	lab1Name := "test-list-1"
	lab2Name := "test-list-2"
	lab1 := createTestLab(lab1Name)
	lab2 := createTestLab(lab2Name)
	defer cleanupTestData(t, storage, lab1Name, lab2Name)

	// Save both
	if err := storage.Save(lab1); err != nil {
		t.Fatalf("failed to save lab1: %v", err)
	}
	if err := storage.Save(lab2); err != nil {
		t.Fatalf("failed to save lab2: %v", err)
	}
	t.Logf("✓ Two labs saved")

	// List all labs
	labs, err := storage.List(false)
	if err != nil {
		t.Fatalf("failed to list labs: %v", err)
	}

	// Verify we got at least our 2 labs (there might be others from previous tests)
	if len(labs) < 2 {
		t.Errorf("expected at least 2 labs, got %d", len(labs))
	}

	// Verify our labs are in the list
	foundLab1 := false
	foundLab2 := false
	for _, lab := range labs {
		if lab.Name == lab1Name {
			foundLab1 = true
		}
		if lab.Name == lab2Name {
			foundLab2 = true
		}
	}

	if !foundLab1 {
		t.Errorf("lab1 %s not found in list", lab1Name)
	}
	if !foundLab2 {
		t.Errorf("lab2 %s not found in list", lab2Name)
	}

	t.Logf("✓ Labs listed successfully: found %d labs", len(labs))
}

func TestDelete(t *testing.T) {
	// Setup
	cfg := getTestConfig()
	storage, err := New(cfg)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}
	defer storage.Close()

	labName := "test-delete"
	lab := createTestLab(labName)

	// Save lab
	if err := storage.Save(lab); err != nil {
		t.Fatalf("failed to save lab: %v", err)
	}
	t.Logf("✓ Lab saved: %s", labName)

	// Verify it exists
	_, err = storage.Get(labName)
	if err != nil {
		t.Fatalf("lab should exist before delete: %v", err)
	}

	// Delete lab
	err = storage.Delete(labName)
	if err != nil {
		t.Fatalf("failed to delete lab: %v", err)
	}
	t.Logf("✓ Lab deleted: %s", labName)

	// Verify it's gone (Get should fail)
	_, err = storage.Get(labName)
	if err == nil {
		t.Errorf("expected error when getting deleted lab, got nil")
	}
	t.Logf("✓ Verified lab is deleted (not retrievable)")
}

func TestGetNonExistent(t *testing.T) {
	// Setup
	cfg := getTestConfig()
	storage, err := New(cfg)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}
	defer storage.Close()

	// Try to get non-existent lab
	_, err = storage.Get("nonexistent-lab-12345")
	if err == nil {
		t.Errorf("expected error when getting non-existent lab, got nil")
	}
	t.Logf("✓ Correctly returned error for non-existent lab: %v", err)
}

func TestSaveUpdate(t *testing.T) {
	// Setup
	cfg := getTestConfig()
	storage, err := New(cfg)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}
	defer storage.Close()

	labName := "test-update"
	lab := createTestLab(labName)
	defer cleanupTestData(t, storage, labName)

	// Save first time
	if err := storage.Save(lab); err != nil {
		t.Fatalf("failed to save lab: %v", err)
	}
	t.Logf("✓ Lab saved initially: %s", labName)

	// Update lab
	lab.Status.State = "stopped"
	lab.ObjectMeta.Labels["updated"] = "true"

	// Save again (should UPSERT)
	if err := storage.Save(lab); err != nil {
		t.Fatalf("failed to update lab: %v", err)
	}
	t.Logf("✓ Lab updated: %s", labName)

	// Retrieve and verify update
	retrieved, err := storage.Get(labName)
	if err != nil {
		t.Fatalf("failed to get updated lab: %v", err)
	}

	if retrieved.Status.State != "stopped" {
		t.Errorf("expected state stopped, got %s", retrieved.Status.State)
	}
	if retrieved.ObjectMeta.Labels["updated"] != "true" {
		t.Errorf("expected updated label, got %s", retrieved.ObjectMeta.Labels["updated"])
	}

	t.Logf("✓ Update verified: state=%s, updated=%s",
		retrieved.Status.State,
		retrieved.ObjectMeta.Labels["updated"])
}
