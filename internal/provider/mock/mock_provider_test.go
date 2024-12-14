package mock

import (
	"errors"
	"testing"

	"github.com/pavelanni/storctl/internal/provider/options"
	"github.com/pavelanni/storctl/internal/types"
)

func TestMockProvider(t *testing.T) {
	t.Run("CreateServer", func(t *testing.T) {
		mock := &MockProvider{
			CreateServerFunc: func(opts options.ServerCreateOpts) (*types.Server, error) {
				if opts.Name == "test-server" {
					return &types.Server{
						ObjectMeta: types.ObjectMeta{
							Name: opts.Name,
						},
					}, nil
				}
				return nil, errors.New("server creation failed")
			},
		}

		// Test successful case
		server, err := mock.CreateServer(options.ServerCreateOpts{Name: "test-server"})
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if server.Name != "test-server" {
			t.Errorf("Expected server name 'test-server', got %s", server.Name)
		}

		// Test error case
		_, err = mock.CreateServer(options.ServerCreateOpts{Name: "invalid-server"})
		if err == nil {
			t.Error("Expected error, got nil")
		}
	})

	t.Run("DeleteServer", func(t *testing.T) {
		mock := &MockProvider{
			DeleteServerFunc: func(name string, force bool) *types.ServerDeleteStatus {
				if name == "test-server" {
					return &types.ServerDeleteStatus{
						Deleted: true,
					}
				}
				return &types.ServerDeleteStatus{
					Error: errors.New("server not found"),
				}
			},
		}

		// Test successful deletion
		status := mock.DeleteServer("test-server", false)
		if !status.Deleted {
			t.Error("Expected server to be deleted")
		}
		if status.Error != nil {
			t.Errorf("Expected no error, got %v", status.Error)
		}

		// Test failed deletion
		status = mock.DeleteServer("nonexistent-server", false)
		if status.Deleted {
			t.Error("Expected server not to be deleted")
		}
		if status.Error == nil {
			t.Error("Expected error, got nil")
		}
	})

	t.Run("GetVolume", func(t *testing.T) {
		expectedVolume := &types.Volume{
			ObjectMeta: types.ObjectMeta{
				Name: "test-volume",
			},
			Spec: types.VolumeSpec{
				Size: 100,
			},
		}

		mock := &MockProvider{
			GetVolumeFunc: func(name string) (*types.Volume, error) {
				if name == "test-volume" {
					return expectedVolume, nil
				}
				return nil, errors.New("volume not found")
			},
		}

		// Test successful case
		volume, err := mock.GetVolume("test-volume")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if volume.Name != expectedVolume.Name {
			t.Errorf("Expected volume name %s, got %s", expectedVolume.Name, volume.Name)
		}
		if volume.Spec.Size != expectedVolume.Spec.Size {
			t.Errorf("Expected volume size %d, got %d", expectedVolume.Spec.Size, volume.Spec.Size)
		}

		// Test error case
		_, err = mock.GetVolume("nonexistent-volume")
		if err == nil {
			t.Error("Expected error, got nil")
		}
	})
}
