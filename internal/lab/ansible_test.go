package lab

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/pavelanni/storctl/internal/config"
	"github.com/pavelanni/storctl/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestManagerSvc_CreateAnsibleInventoryFile(t *testing.T) {
	// Use a fixed directory instead of a temporary one
	tmpDir := filepath.Join(os.Getenv("HOME"), "storctl-test")
	t.Logf("Creating directory structure in: %s", tmpDir)

	// Create necessary subdirectories
	fullPath := filepath.Join(tmpDir, config.DefaultConfigDir, config.DefaultAnsibleDir)
	t.Logf("Full path to create: %s", fullPath)

	err := os.MkdirAll(fullPath, 0755)
	assert.NoError(t, err)

	// Verify directory exists after creation
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		t.Errorf("Directory was not created: %s", fullPath)
	} else {
		t.Logf("Directory successfully created: %s", fullPath)
	}

	// Mock the home directory by setting up a test environment variable
	os.Setenv("HOME", tmpDir)

	// Create a test lab instance
	lab := &types.Lab{
		ObjectMeta: types.ObjectMeta{
			Name: "test-lab",
		},
		Spec: types.LabSpec{
			LetsEncrypt: "staging",
			CertManager: true,
		},
		Status: types.LabStatus{
			Servers: []*types.Server{
				{
					ObjectMeta: types.ObjectMeta{
						Name: "lab1-cp",
					},
					Status: types.ServerStatus{
						PublicNet: &types.PublicNet{
							IPv4: &struct {
								IP string `json:"ip"`
							}{
								IP: "192.168.1.10",
							},
						},
					},
				},
				{
					ObjectMeta: types.ObjectMeta{
						Name: "lab1-worker-1",
					},
					Status: types.ServerStatus{
						PublicNet: &types.PublicNet{
							IPv4: &struct {
								IP string `json:"ip"`
							}{
								IP: "192.168.1.20",
							},
						},
					},
				},
			},
		},
	}

	// Create manager service instance
	m := &ManagerSvc{}

	// Test the function
	err = m.CreateAnsibleInventoryFile(lab)
	assert.NoError(t, err)

	// Verify the file was created
	inventoryFile := filepath.Join(tmpDir, config.DefaultConfigDir, config.DefaultAnsibleDir, "test-lab-inventory.json")
	assert.FileExists(t, inventoryFile)

	// Read and parse the created file
	data, err := os.ReadFile(inventoryFile)
	assert.NoError(t, err)

	var inventory Inventory
	err = json.Unmarshal(data, &inventory)
	assert.NoError(t, err)

	// Verify the inventory structure
	assert.Contains(t, inventory.All.Children, "control_plane")
	assert.Contains(t, inventory.All.Children, "nodes")

	// Verify control plane hosts
	assert.Contains(t, inventory.All.Children["controlplane"].Hosts, "lab1-cp")
	assert.Equal(t, "192.168.1.10", inventory.All.Children["controlplane"].Hosts["lab1-cp"].AnsibleHost)

	// Verify worker nodes
	assert.Contains(t, inventory.All.Children["nodes"].Hosts, "lab1-worker-1")
	assert.Equal(t, "192.168.1.20", inventory.All.Children["nodes"].Hosts["lab1-worker-1"].AnsibleHost)

	// Verify variables
	assert.Equal(t, config.DefaultAdminUser, inventory.All.Vars["ansible_user"])
	assert.Equal(t, "staging", inventory.All.Vars["letsencrypt_environment"])
	assert.Equal(t, true, inventory.All.Vars["cert_manager_enable"])
	assert.Equal(t, "test-lab", inventory.All.Vars["lab_name"])

	// Add verification at the end of the test
	t.Logf("Checking if directory still exists at end of test")
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		t.Errorf("Directory no longer exists at end of test: %s", fullPath)
	}
}
