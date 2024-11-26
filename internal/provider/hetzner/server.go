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

func (p *HetznerProvider) CreateServer(name string, serverType string, image string, location string, sshKeyNames []string, labels map[string]string) (*types.Server, error) {
	// DEBUG
	fmt.Printf("Creating server %s with type %s, image %s, location %s, ssh keys %v\n", name, serverType, image, location, sshKeyNames)
	hCloudSSHKeys := make([]*hcloud.SSHKey, 0)
	for _, sshKeyName := range sshKeyNames {
		sshKey, _, err := p.Client.SSHKey.GetByName(context.Background(), sshKeyName)
		if err != nil {
			p.logger.Info("SSH key not found, skipping",
				"key", sshKeyName)
			continue
		}
		hCloudSSHKeys = append(hCloudSSHKeys, sshKey)
	}
	if len(hCloudSSHKeys) == 0 {
		return nil, fmt.Errorf("no SSH keys provided")
	}
	p.logger.Info("creating server",
		"name", name,
		"type", serverType,
		"image", image,
		"location", location,
		"ssh_keys", sshKeyNames)
	server, _, err := p.Client.Server.Create(context.Background(), hcloud.ServerCreateOpts{
		Name: name,
		ServerType: &hcloud.ServerType{
			Name: serverType,
		},
		Image: &hcloud.Image{
			Name: image,
		},
		Location: &hcloud.Location{
			Name: location,
		},
		SSHKeys: hCloudSSHKeys,
		Labels:  labels,
	})
	if err != nil {
		return nil, err
	}
	p.logger.Info("successfully created server",
		"name", name,
		"ip", server.Server.PublicNet.IPv4.IP)
	return mapServer(server.Server, p.Client), nil
}
func (p *HetznerProvider) GetServer(serverName string) (*types.Server, error) {
	server, _, err := p.Client.Server.Get(context.Background(), serverName)
	if err != nil {
		return nil, err
	}
	return mapServer(server, p.Client), nil
}

func (p *HetznerProvider) AllServers() ([]*types.Server, error) {
	servers, err := p.Client.Server.All(context.Background())
	if err != nil {
		return nil, err
	}

	return mapServers(servers, p.Client), nil
}

func (p *HetznerProvider) DeleteServer(serverName string, force bool) error {
	if serverName == "" {
		return fmt.Errorf("empty server name provided")
	}
	server, _, err := p.Client.Server.Get(context.Background(), serverName)
	if err != nil {
		return err
	}
	if server == nil {
		p.logger.Info("Server not found, skipping",
			"server", serverName)
		return nil
	}
	if !force {
		if deleteAfterStr, ok := server.Labels["delete_after"]; ok {
			deleteAfter, err := time.Parse(time.RFC3339, deleteAfterStr)
			if err == nil && time.Now().Before(deleteAfter) {
				p.logger.Warn("server not ready for deletion",
					"server", serverName,
					"delete_after", deleteAfter.Format("2006-01-02 15:04:05"))
				return fmt.Errorf("server %s is not ready for deletion until %s", serverName, deleteAfter)
			}
		}
	}
	for _, volume := range server.Volumes {
		_, _, err = p.Client.Volume.Detach(context.Background(), volume)
		if err != nil {
			return err
		}
	}
	_, _, err = p.Client.Server.DeleteWithResult(context.Background(), server)
	return err
}

// mapServer converts a Hetzner-specific server to our generic Server type
func mapServer(s *hcloud.Server, client *hcloud.Client) *types.Server {
	if s == nil {
		return nil
	}

	return &types.Server{
		ID:          strconv.FormatInt(s.ID, 10),
		Name:        s.Name,
		Status:      string(s.Status),
		Type:        s.ServerType.Name,
		Owner:       s.Labels["owner"],
		Cores:       s.ServerType.Cores,
		Memory:      float32(s.ServerType.Memory),
		Disk:        s.ServerType.Disk,
		Location:    s.Datacenter.Location.Name,
		Labels:      s.Labels,
		Volumes:     mapVolumes(s.Volumes, client),
		Created:     s.Created,
		DeleteAfter: timeutil.ParseDeleteAfter(s.Labels["delete_after"]),
	}
}

// mapServers converts a slice of Hetzner servers
func mapServers(servers []*hcloud.Server, client *hcloud.Client) []*types.Server {
	if servers == nil {
		return nil
	}

	result := make([]*types.Server, len(servers))
	for i, s := range servers {
		result[i] = mapServer(s, client)
	}
	return result
}
