package hetzner

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/pavelanni/storctl/internal/provider/options"
	"github.com/pavelanni/storctl/internal/types"
	"github.com/pavelanni/storctl/internal/util/timeutil"
)

func (p *HetznerProvider) CreateVolume(opts options.VolumeCreateOpts) (*types.Volume, error) {
	var hCloudServer *hcloud.Server
	var hCloudLocation *hcloud.Location
	var err error
	if opts.ServerName == "" {
		if opts.Location == "" {
			return nil, fmt.Errorf("location is required when server name is empty")
		}
		p.logger.Debug("server name is empty, using location instead",
			"location", opts.Location)
	} else {
		hCloudServer, _, err = p.Client.Server.GetByName(context.Background(), opts.ServerName)
		if err != nil {
			return nil, fmt.Errorf("error getting server: %w", err)
		}
		if hCloudServer == nil {
			return nil, fmt.Errorf("server not found: %s", opts.ServerName)
		}
		location := hCloudServer.Datacenter.Location.Name
		if location == "" {
			return nil, fmt.Errorf("server %s has no location", opts.ServerName)
		}
		hCloudLocation, _, err = p.Client.Location.GetByName(context.Background(), location)
		if err != nil {
			return nil, fmt.Errorf("error getting location: %w", err)
		}
		if hCloudLocation == nil {
			return nil, fmt.Errorf("location not found: %s", location)
		}
	}
	volumeOpts := hcloud.VolumeCreateOpts{
		Name:      opts.Name,
		Size:      opts.Size,
		Server:    hCloudServer,
		Labels:    opts.Labels,
		Automount: &opts.Automount,
		Format:    &opts.Format,
	}
	p.logger.Debug("creating volume",
		"name", volumeOpts.Name,
		"size", volumeOpts.Size,
		"server", volumeOpts.Server.Name,
		"automount", *volumeOpts.Automount,
		"format", *volumeOpts.Format)

	volume, _, err := p.Client.Volume.Create(context.Background(), volumeOpts)
	if err != nil {
		return nil, fmt.Errorf("error creating volume: %w", err)
	}
	p.logger.Debug("successfully created volume",
		"name", volumeOpts.Name)
	return p.mapVolume(volume.Volume), nil
}

func (p *HetznerProvider) AttachVolume(volumeName, serverName string) error {
	volume, _, err := p.Client.Volume.Get(context.Background(), volumeName)
	if err != nil {
		return fmt.Errorf("error getting volume: %w", err)
	}
	if volume == nil {
		return fmt.Errorf("volume not found: %s", volumeName)
	}
	if volume.Server != nil {
		return fmt.Errorf("volume already attached to server: %s", volume.Server.Name)
	}
	hCloudServer, _, err := p.Client.Server.GetByName(context.Background(), serverName)
	if err != nil {
		return fmt.Errorf("error getting server: %w", err)
	}
	if hCloudServer == nil {
		return fmt.Errorf("server not found: %s", serverName)
	}
	_, _, err = p.Client.Volume.Attach(context.Background(), volume, hCloudServer)
	return fmt.Errorf("error attaching volume: %w", err)
}

func (p *HetznerProvider) GetVolume(volumeName string) (*types.Volume, error) {
	volume, _, err := p.Client.Volume.Get(context.Background(), volumeName)
	if err != nil {
		return nil, fmt.Errorf("error getting volume: %w", err)
	}
	return p.mapVolume(volume), nil
}

func (p *HetznerProvider) ListVolumes(opts options.VolumeListOpts) ([]*types.Volume, error) {
	volumes, _, err := p.Client.Volume.List(context.Background(), hcloud.VolumeListOpts{
		ListOpts: hcloud.ListOpts{
			LabelSelector: opts.LabelSelector,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("error listing volumes: %w", err)
	}
	return p.mapVolumes(volumes), nil
}

func (p *HetznerProvider) AllVolumes() ([]*types.Volume, error) {
	volumes, err := p.Client.Volume.All(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error listing volumes: %w", err)
	}

	return p.mapVolumes(volumes), nil
}

func (p *HetznerProvider) DeleteVolume(volumeName string, force bool) *types.VolumeDeleteStatus {
	if volumeName == "" {
		return &types.VolumeDeleteStatus{
			Error: fmt.Errorf("empty volume name provided"),
		}
	}
	volume, _, err := p.Client.Volume.Get(context.Background(), volumeName)
	if err != nil {
		return &types.VolumeDeleteStatus{
			Error: fmt.Errorf("error getting volume: %w", err),
		}
	}
	if volume == nil {
		p.logger.Debug("Volume not found, skipping",
			"volume", volumeName)
		return &types.VolumeDeleteStatus{
			Deleted: true,
		}
	}
	if !force {
		if deleteAfterStr, ok := volume.Labels["delete_after"]; ok {
			deleteAfter, err := time.Parse("2006-01-02-15-04", deleteAfterStr)
			if err != nil {
				p.logger.Error("failed to parse delete_after label",
					"volume", volumeName,
					"delete_after", deleteAfterStr,
					"error", err)
				return &types.VolumeDeleteStatus{
					Error: fmt.Errorf("error parsing delete_after label: %w", err),
				}
			}
			if time.Now().UTC().Before(deleteAfter) {
				p.logger.Warn("volume not ready for deletion",
					"volume", volumeName,
					"delete_after", deleteAfter.Format("2006-01-02 15:04:05"))
				return &types.VolumeDeleteStatus{
					DeleteAfter: deleteAfter,
				}
			}
		}
	}
	// check if volume is attached to any server
	if volume.Server != nil {
		p.logger.Debug("volume is attached to server, detaching",
			"volume", volume.Name,
			"server", volume.Server.Name)
		_, _, err = p.Client.Volume.Detach(context.Background(), volume)
		if err != nil {
			return &types.VolumeDeleteStatus{
				Error: fmt.Errorf("error detaching volume: %w", err),
			}
		}
	}
	_, err = p.Client.Volume.Delete(context.Background(), volume)
	if err != nil {
		return &types.VolumeDeleteStatus{
			Error: fmt.Errorf("error deleting volume: %w", err),
		}
	}
	return &types.VolumeDeleteStatus{
		Deleted: true,
	}
}

// mapVolume converts a Hetzner-specific volume to our generic Volume type
func (p *HetznerProvider) mapVolume(v *hcloud.Volume) *types.Volume {
	if v == nil {
		return nil
	}

	// Handle pointer fields safely
	var location string
	if v.Location != nil {
		location = v.Location.Name
	}

	var format string
	if v.Format != nil {
		format = *v.Format
	}

	var serverID int64
	var serverName string
	if v.Server != nil {
		serverID = int64(v.Server.ID)
		// Fetch server details from Hetzner
		if server, _, err := p.Client.Server.GetByID(context.Background(), v.Server.ID); err == nil && server != nil {
			serverName = server.Name
		}
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
			Location:   location,
			Size:       v.Size,
			Format:     format,
			Labels:     v.Labels,
			ServerID:   strconv.FormatInt(serverID, 10),
			ServerName: serverName,
		},
		Status: types.VolumeStatus{
			Status:      string(v.Status),
			Owner:       v.Labels["owner"],
			Created:     v.Created,
			DeleteAfter: timeutil.ParseDeleteAfter(v.Labels["delete_after"]),
		},
	}
}

// mapVolumes converts a slice of Hetzner volumes
func (p *HetznerProvider) mapVolumes(volumes []*hcloud.Volume) []*types.Volume {
	if volumes == nil {
		return nil
	}

	result := make([]*types.Volume, len(volumes))
	for i, v := range volumes {
		result[i] = p.mapVolume(v)
	}
	return result
}
