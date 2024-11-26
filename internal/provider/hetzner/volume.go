package hetzner

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/pavelanni/labshop/internal/types"
	"github.com/pavelanni/labshop/internal/util/timeutil"
)

func (p *HetznerProvider) CreateVolume(name string, size int, server string, labels map[string]string, automount bool, format string) (*types.Volume, error) {
	hCloudServer := &hcloud.Server{}
	var err error
	if server != "" {
		hCloudServer, _, err = p.Client.Server.GetByName(context.Background(), server)
		if err != nil {
			return nil, err
		}
	}
	p.logger.Info("creating volume",
		"name", name,
		"size", size,
		"server", server,
		"automount", automount,
		"format", format)
	volume, _, err := p.Client.Volume.Create(context.Background(), hcloud.VolumeCreateOpts{
		Name:      name,
		Size:      size,
		Server:    hCloudServer,
		Automount: &automount,
		Format:    &format,
	})
	if err != nil {
		return nil, err
	}
	p.logger.Info("successfully created volume",
		"name", name)
	return mapVolume(volume.Volume, p.Client), nil
}

func (p *HetznerProvider) GetVolume(volumeName string) (*types.Volume, error) {
	volume, _, err := p.Client.Volume.Get(context.Background(), volumeName)
	if err != nil {
		return nil, err
	}
	return mapVolume(volume, p.Client), nil
}

func (p *HetznerProvider) AllVolumes() ([]*types.Volume, error) {
	volumes, err := p.Client.Volume.All(context.Background())
	if err != nil {
		return nil, err
	}

	return mapVolumes(volumes, p.Client), nil
}

func (p *HetznerProvider) DeleteVolume(volumeName string, force bool) error {
	if volumeName == "" {
		return fmt.Errorf("empty volume name provided")
	}
	volume, _, err := p.Client.Volume.Get(context.Background(), volumeName)
	if err != nil {
		return err
	}
	if volume == nil {
		p.logger.Info("Volume not found, skipping",
			"volume", volumeName)
		return nil
	}
	if !force {
		if deleteAfterStr, ok := volume.Labels["delete_after"]; ok {
			deleteAfter, err := time.Parse(time.RFC3339, deleteAfterStr)
			if err == nil && time.Now().Before(deleteAfter) {
				p.logger.Warn("volume not ready for deletion",
					"volume", volumeName,
					"delete_after", deleteAfter.Format("2006-01-02 15:04:05"))
				return fmt.Errorf("volume %s is not ready for deletion until %s", volumeName, deleteAfter)
			}
		}
	}
	// check if volume is attached to any server
	if volume.Server != nil {
		p.logger.Info("volume is attached to server, detaching",
			"volume", volume.Name,
			"server", volume.Server.Name)
		_, _, err = p.Client.Volume.Detach(context.Background(), volume)
		if err != nil {
			return err
		}
	}
	_, err = p.Client.Volume.Delete(context.Background(), volume)
	return err
}

// mapVolume converts a Hetzner-specific volume to our generic Volume type
func mapVolume(v *hcloud.Volume, client *hcloud.Client) *types.Volume {
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
		if server, _, err := client.Server.GetByID(context.Background(), v.Server.ID); err == nil && server != nil {
			serverName = server.Name
		}
	}

	return &types.Volume{
		ID:          strconv.FormatInt(v.ID, 10),
		Name:        v.Name,
		Status:      string(v.Status),
		Owner:       v.Labels["owner"],
		ServerID:    strconv.FormatInt(serverID, 10),
		ServerName:  serverName,
		Location:    location,
		Size:        v.Size,
		Format:      format,
		Labels:      v.Labels,
		Created:     v.Created,
		DeleteAfter: timeutil.ParseDeleteAfter(v.Labels["delete_after"]),
	}
}

// mapVolumes converts a slice of Hetzner volumes
func mapVolumes(volumes []*hcloud.Volume, client *hcloud.Client) []*types.Volume {
	if volumes == nil {
		return nil
	}

	result := make([]*types.Volume, len(volumes))
	for i, v := range volumes {
		result[i] = mapVolume(v, client)
	}
	return result
}
