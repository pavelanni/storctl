package lima

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/pavelanni/storctl/internal/provider/options"
	"github.com/stretchr/testify/assert"
)

func checkLimaCtl(t *testing.T) {
	_, err := exec.LookPath("limactl")
	if err != nil {
		t.Skip("limactl not found in PATH, skipping test")
	}
}

func TestCreateServer(t *testing.T) {
	checkLimaCtl(t)

	provider := &LimaProvider{
		arch: "arm64",
	}

	// Test case 1: Valid server creation
	t.Run("valid server creation", func(t *testing.T) {
		opts := options.ServerCreateOpts{
			Name:  "test-server",
			Type:  "cx22",
			Image: "ubuntu-22.04",
		}
		// cleanup before test
		_ = provider.DeleteServer("test-server", true)
		server, err := provider.CreateServer(opts)
		assert.NoError(t, err)
		assert.NotNil(t, server)

	})

	// Test case 1.1: Server already exists
	t.Run("server already exists", func(t *testing.T) {
		opts := options.ServerCreateOpts{
			Name:  "test-server",
			Type:  "cx22",
			Image: "ubuntu-22.04",
		}
		_, err := provider.CreateServer(opts)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
		// Cleanup
		_ = provider.DeleteServer("test-server", true)
	})

	// Test case 2: Invalid server type
	t.Run("invalid server type", func(t *testing.T) {
		opts := options.ServerCreateOpts{
			Name:  "test-server-invalid-type",
			Type:  "invalid-type",
			Image: "ubuntu-22.04",
		}

		server, err := provider.CreateServer(opts)
		assert.Error(t, err)
		assert.Nil(t, server)
		assert.Contains(t, err.Error(), "invalid server type")
	})

	// Test case 3: Server with additional disks
	t.Run("server with additional disks", func(t *testing.T) {
		opts := options.ServerCreateOpts{
			Name:  "test-server-disks",
			Type:  "cx22",
			Image: "ubuntu-22.04",
			AdditionalDisks: []options.AdditionalDisk{
				{
					Name:   "data",
					Format: true,
					FsType: "ext4",
				},
			},
		}
		// cleanup before test
		_ = provider.DeleteServer("test-server-disks", true)
		server, err := provider.CreateServer(opts)
		assert.NoError(t, err)
		assert.NotNil(t, server)

		// Cleanup
		_ = provider.DeleteServer("test-server-disks", true)
	})
}

func TestCreateLimaConfig(t *testing.T) {
	tests := []struct {
		name     string
		server   ConfigServer
		expected LimaConfig
	}{
		{
			name: "basic ubuntu server",
			server: ConfigServer{
				Name:   "test-server",
				CPUs:   2,
				Memory: "4GB",
				Disk:   "40GB",
				Image:  "ubuntu-22.04",
				Arch:   "arm64",
			},
			expected: LimaConfig{
				VMType: "qemu",
				CPUs:   2,
				Memory: "4GB",
				Disk:   "40GB",
				Arch:   "aarch64",
				OS:     "Linux",
				Images: []ConfigImage{
					{
						Location: "https://cloud-images.ubuntu.com/releases/22.04/release/ubuntu-22.04-server-cloudimg-arm64.img",
						Arch:     "aarch64",
					},
				},
				Networks: []ConfigNetwork{
					{
						LimaNetwork: "shared",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := createLimaConfig(tt.server)
			assert.Equal(t, tt.expected, config)
		})
	}
}

func TestWriteConfig(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test-config.yaml")

	config := LimaConfig{
		VMType: "qemu",
		CPUs:   2,
		Memory: "4GB",
		Disk:   "40GB",
	}

	err := writeConfig(configFile, config)
	assert.NoError(t, err)

	// Verify file exists and has correct permissions
	info, err := os.Stat(configFile)
	assert.NoError(t, err)
	assert.Equal(t, os.FileMode(0644), info.Mode().Perm())
}

func TestGetArchConversions(t *testing.T) {
	// Test getArchForImage
	t.Run("getArchForImage", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{"amd64", "amd64"},
			{"x86_64", "amd64"},
			{"arm64", "arm64"},
			{"aarch64", "arm64"},
			{"unknown", "arm64"},
		}

		for _, tt := range tests {
			result := getArchForImage(tt.input)
			assert.Equal(t, tt.expected, result)
		}
	})

	// Test getArchForArch
	t.Run("getArchForArch", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{"amd64", "x86_64"},
			{"x86_64", "x86_64"},
			{"arm64", "aarch64"},
			{"aarch64", "aarch64"},
			{"unknown", "aarch64"},
		}

		for _, tt := range tests {
			result := getArchForArch(tt.input)
			assert.Equal(t, tt.expected, result)
		}
	})
}
