package lima

import (
	"os/exec"
	"testing"

	"github.com/pavelanni/storctl/internal/provider/options"
	"github.com/pavelanni/storctl/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestCreateVolume(t *testing.T) {
	if _, err := exec.LookPath("limactl"); err != nil {
		t.Skip("limactl not found in PATH, skipping test")
	}

	tests := []struct {
		name    string
		opts    options.VolumeCreateOpts
		wantErr bool
	}{
		{
			name: "create valid volume",
			opts: options.VolumeCreateOpts{
				Name: "test-volume",
				Size: 10,
			},
			wantErr: false,
		},
		{
			name: "create volume with empty name",
			opts: options.VolumeCreateOpts{
				Name: "",
				Size: 10,
			},
			wantErr: true,
		},
	}

	provider := &LimaProvider{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			volume, err := provider.CreateVolume(tt.opts)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, volume)
			assert.Equal(t, tt.opts.Name, volume.Name)
			assert.Equal(t, tt.opts.Size, volume.Spec.Size)
		})
	}
}

func TestGetVolume(t *testing.T) {
	if _, err := exec.LookPath("limactl"); err != nil {
		t.Skip("limactl not found in PATH, skipping test")
	}

	tests := []struct {
		name    string
		volName string
		wantErr bool
	}{
		{
			name:    "get existing volume",
			volName: "test-volume",
			wantErr: false,
		},
		{
			name:    "get non-existent volume",
			volName: "non-existent-volume",
			wantErr: true,
		},
	}

	provider := &LimaProvider{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			volume, err := provider.GetVolume(tt.volName)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, volume)
			assert.Equal(t, tt.volName, volume.Name)
		})
	}
}

func TestListVolumes(t *testing.T) {
	if _, err := exec.LookPath("limactl"); err != nil {
		t.Skip("limactl not found in PATH, skipping test")
	}

	provider := &LimaProvider{}
	volumes, err := provider.ListVolumes(options.VolumeListOpts{})
	assert.NoError(t, err)
	assert.NotNil(t, volumes)
}

func TestDeleteVolume(t *testing.T) {
	if _, err := exec.LookPath("limactl"); err != nil {
		t.Skip("limactl not found in PATH, skipping test")
	}

	tests := []struct {
		name      string
		volName   string
		force     bool
		wantError bool
	}{
		{
			name:      "delete existing volume",
			volName:   "test-volume",
			force:     false,
			wantError: false,
		},
		{
			name:      "delete non-existent volume",
			volName:   "non-existent-volume",
			force:     false,
			wantError: false,
		},
	}

	provider := &LimaProvider{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := provider.DeleteVolume(tt.volName, tt.force)
			if tt.wantError {
				assert.False(t, status.Deleted)
				assert.Error(t, status.Error)
				return
			}
			assert.True(t, status.Deleted)
			assert.Nil(t, status.Error)
		})
	}
}

func TestMapVolume(t *testing.T) {
	if _, err := exec.LookPath("limactl"); err != nil {
		t.Skip("limactl not found in PATH, skipping test")
	}

	tests := []struct {
		name     string
		limaDisk *ConfigDisk
		want     *types.Volume
	}{
		{
			name: "map valid disk",
			limaDisk: &ConfigDisk{
				Name:     "test-disk",
				Size:     10737418240, // 10 GiB in bytes
				Instance: "test-instance",
			},
			want: &types.Volume{
				TypeMeta: types.TypeMeta{
					APIVersion: "v1",
					Kind:       "Volume",
				},
				ObjectMeta: types.ObjectMeta{
					Name: "test-disk",
				},
				Spec: types.VolumeSpec{
					Size:       10,
					ServerName: "test-instance",
					Provider:   "lima",
				},
			},
		},
		{
			name:     "map nil disk",
			limaDisk: nil,
			want:     nil,
		},
	}

	provider := &LimaProvider{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := provider.mapVolume(tt.limaDisk)
			assert.Equal(t, tt.want, got)
		})
	}
}
