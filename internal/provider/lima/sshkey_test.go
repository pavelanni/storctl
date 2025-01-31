package lima

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pavelanni/storctl/internal/provider/options"
	"github.com/stretchr/testify/assert"
)

// Add this variable at package level
var userHomeDir = os.UserHomeDir

func setupTestEnvironment(t *testing.T) (*LimaProvider, string, func()) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "lima-test")
	if err != nil {
		t.Fatal(err)
	}

	// Create the _config directory structure
	configDir := filepath.Join(tmpDir, ".lima", "_config")
	err = os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	// Create a mock SSH public key
	mockPublicKey := "ssh-rsa AAAAB3NzaC1yc2EA... test@example.com"
	err = os.WriteFile(filepath.Join(configDir, "user.pub"), []byte(mockPublicKey), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Create cleanup function
	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	provider := &LimaProvider{}
	return provider, tmpDir, cleanup
}

func TestGetSSHKey(t *testing.T) {
	provider, tmpDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Store original and override userHomeDir
	originalUserHomeDir := userHomeDir
	userHomeDir = func() (string, error) {
		return tmpDir, nil
	}
	defer func() {
		userHomeDir = originalUserHomeDir
	}()

	// Test getting the default key
	key, err := provider.GetSSHKey("default")
	assert.NoError(t, err)
	assert.NotNil(t, key)
	assert.Equal(t, "default", key.Name)
	assert.Contains(t, key.Spec.PublicKey, "ssh-ed25519")
}

func TestListSSHKeys(t *testing.T) {
	provider, tmpDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Store original and override userHomeDir
	originalUserHomeDir := userHomeDir
	userHomeDir = func() (string, error) {
		return tmpDir, nil
	}
	defer func() {
		userHomeDir = originalUserHomeDir
	}()

	// Test listing keys
	keys, err := provider.ListSSHKeys(options.SSHKeyListOpts{})
	assert.NoError(t, err)
	assert.Len(t, keys, 1)
	assert.Equal(t, "default", keys[0].Name)
}

func TestCreateSSHKey(t *testing.T) {
	provider, tmpDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Store original and override userHomeDir
	originalUserHomeDir := userHomeDir
	userHomeDir = func() (string, error) {
		return tmpDir, nil
	}
	defer func() {
		userHomeDir = originalUserHomeDir
	}()

	// Test creating a key (should return the default key)
	key, err := provider.CreateSSHKey(options.SSHKeyCreateOpts{
		Name: "test-key",
	})
	assert.NoError(t, err)
	assert.NotNil(t, key)
	assert.Equal(t, "default", key.Name)
}

func TestDeleteSSHKey(t *testing.T) {
	provider := &LimaProvider{}

	// Test deleting a key (should always return not deleted)
	status := provider.DeleteSSHKey("test-key", false)
	assert.NotNil(t, status)
	assert.False(t, status.Deleted)
}

func TestCloudKeyExists(t *testing.T) {
	provider := &LimaProvider{}

	// Test checking if key exists (should always return true)
	exists, err := provider.CloudKeyExists("any-key")
	assert.NoError(t, err)
	assert.True(t, exists)
}

func TestKeyNamesToSSHKeys(t *testing.T) {
	provider, tmpDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Store original and override userHomeDir
	originalUserHomeDir := userHomeDir
	userHomeDir = func() (string, error) {
		return tmpDir, nil
	}
	defer func() {
		userHomeDir = originalUserHomeDir
	}()

	// Test converting key names to SSH keys
	keys, err := provider.KeyNamesToSSHKeys([]string{"key1", "key2"}, options.SSHKeyCreateOpts{})
	assert.NoError(t, err)
	assert.Len(t, keys, 1)
	assert.Equal(t, "default", keys[0].Name)
}
