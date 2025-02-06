package lima

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/pavelanni/storctl/internal/provider/options"
	"github.com/pavelanni/storctl/internal/types"
)

func (p *LimaProvider) CreateVolume(opts options.VolumeCreateOpts) (*types.Volume, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if opts.Name == "" {
		return nil, fmt.Errorf("volume name is required")
	}
	if opts.Size == 0 {
		return nil, fmt.Errorf("volume size is required")
	}
	diskName := opts.Name
	sizeStr := fmt.Sprintf("%dGiB", opts.Size)
	err := createDisk(ctx, diskName, sizeStr)
	if err != nil {
		return nil, fmt.Errorf("error creating disk: %w", err)
	}
	volume := &types.Volume{
		ObjectMeta: types.ObjectMeta{
			Name: diskName,
		},
		Spec: types.VolumeSpec{
			Size: opts.Size,
		},
	}
	return volume, nil
}

func (p *LimaProvider) GetVolume(name string) (*types.Volume, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	listCmd := exec.CommandContext(ctx, "limactl", "disk", "list", "--json", name)
	output, err := listCmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error listing disks: %w, output: %s", err, output)
	}
	disk := &ConfigDisk{}
	err = json.Unmarshal(output, disk)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling disk: %w", err)
	}
	return p.mapVolume(disk), nil
}

func (p *LimaProvider) ListVolumes(opts options.VolumeListOpts) ([]*types.Volume, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var labName string
	if opts.ListOpts.LabelSelector != "" {
		label := opts.ListOpts.LabelSelector
		labName = strings.TrimPrefix(label, "lab_name=")
	}
	listCmd := exec.CommandContext(ctx, "limactl", "disk", "list", "--json")
	output, err := listCmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error listing disks: %w, output: %s", err, output)
	}
	disks := []*ConfigDisk{}
	for _, line := range strings.Split(string(output), "\n") {
		if line == "" {
			continue
		}
		disk := &ConfigDisk{}
		err = json.Unmarshal([]byte(line), disk)
		if err != nil {
			return nil, fmt.Errorf("error unmarshalling disk %s: %w", line, err)
		}
		if strings.HasPrefix(disk.Name, labName) {
			disks = append(disks, disk)
		}
	}
	return p.mapVolumes(disks), nil
}

func (p *LimaProvider) AllVolumes() ([]*types.Volume, error) {
	return p.ListVolumes(options.VolumeListOpts{})
}

func (p *LimaProvider) DeleteVolume(name string, force bool) *types.VolumeDeleteStatus {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	deleteCmd := exec.CommandContext(ctx, "limactl", "disk", "delete", name)
	output, err := deleteCmd.CombinedOutput()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return &types.VolumeDeleteStatus{
				Deleted: false,
				Error:   fmt.Errorf("timeout while deleting disk: %w", err),
			}
		}
		return &types.VolumeDeleteStatus{
			Deleted: false,
			Error:   fmt.Errorf("error deleting disk: %w, output: %s", err, output),
		}
	}
	return &types.VolumeDeleteStatus{Deleted: true}
}

// createDisk creates a disk using limactl command
func createDisk(ctx context.Context, diskName, size string) error {
	// Check if disk already exists using limactl disk list
	listCmd := exec.CommandContext(ctx, "limactl", "disk", "list")
	output, err := listCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error listing disks: %w, output: %s", err, output)
	}

	// If disk is in the list, skip creation
	if strings.Contains(string(output), diskName) {
		fmt.Printf("Disk %s already exists, skipping creation\n", diskName)
		return nil
	}

	// Create disk using limactl disk create
	createCmd := exec.CommandContext(ctx, "limactl", "disk", "create", "--size", size, diskName)
	output, err = createCmd.CombinedOutput()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("timeout while creating disk: %w", err)
		}
		return fmt.Errorf("error creating disk: %w, output: %s", err, output)
	}

	fmt.Printf("Created disk %s of size %s\n", diskName, size)
	return nil
}

func (p *LimaProvider) mapVolume(v *ConfigDisk) *types.Volume {
	if v == nil {
		return nil
	}
	return &types.Volume{
		TypeMeta: types.TypeMeta{
			APIVersion: "v1",
			Kind:       "Volume",
		},
		ObjectMeta: types.ObjectMeta{
			Name: v.Name,
		},
		Spec: types.VolumeSpec{
			Size:       v.Size / 1024 / 1024 / 1024, // convert to GiB
			ServerName: v.Instance,
			Provider:   "lima",
		},
	}
}

// mapVolumes converts a slice of Lima disks to a slice of volumes
func (p *LimaProvider) mapVolumes(disks []*ConfigDisk) []*types.Volume {
	if disks == nil {
		return nil
	}
	result := make([]*types.Volume, len(disks))
	for i, v := range disks {
		result[i] = p.mapVolume(v)
	}
	return result
}
