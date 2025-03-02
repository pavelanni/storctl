package hetzner

import (
	"context"
	"fmt"
	"time"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/pavelanni/storctl/internal/config"
	"github.com/pavelanni/storctl/internal/provider/options"
	"github.com/pavelanni/storctl/internal/types"
	"github.com/pavelanni/storctl/internal/util/timeutil"
)

func (p *HetznerProvider) CreateServer(opts options.ServerCreateOpts) (*types.Server, error) {
	serverOpts := hcloud.ServerCreateOpts{
		Name:       opts.Name,
		ServerType: &hcloud.ServerType{Name: opts.Type},
		Image:      &hcloud.Image{Name: opts.Image},
		Location:   &hcloud.Location{Name: opts.Location},
		Labels:     opts.Labels,
		UserData:   opts.UserData,
	}
	sshKeyNames := make([]string, 0)
	for _, sshKey := range opts.SSHKeys {
		sshKeyNames = append(sshKeyNames, sshKey.ObjectMeta.Name)
	}
	hCloudSSHKeys := make([]*hcloud.SSHKey, 0)
	for _, sshKey := range opts.SSHKeys {
		hCloudKey, _, err := p.Client.SSHKey.Get(context.Background(), sshKey.ObjectMeta.Name)
		if err != nil {
			return nil, fmt.Errorf("error getting SSH key: %w", err)
		}
		hCloudSSHKeys = append(hCloudSSHKeys, hCloudKey)
	}
	if len(hCloudSSHKeys) == 0 {
		return nil, fmt.Errorf("no SSH keys provided")
	}
	serverOpts.SSHKeys = hCloudSSHKeys
	p.logger.Debug("creating server",
		"name", opts.Name,
		"type", opts.Type,
		"image", opts.Image,
		"location", opts.Location,
		"ssh_keys", sshKeyNames)
	server, _, err := p.Client.Server.Create(context.Background(), serverOpts)
	if err != nil {
		return nil, fmt.Errorf("error creating server: %w", err)
	}
	p.logger.Debug("successfully created server",
		"name", opts.Name,
		"ip", server.Server.PublicNet.IPv4.IP)

	return p.mapServer(server.Server), nil
}

func (p *HetznerProvider) GetServer(serverName string) (*types.Server, error) {
	server, _, err := p.Client.Server.Get(context.Background(), serverName)
	if err != nil {
		return nil, fmt.Errorf("error getting server: %w", err)
	}
	return p.mapServer(server), nil
}

func (p *HetznerProvider) ListServers(opts options.ServerListOpts) ([]*types.Server, error) {
	servers, _, err := p.Client.Server.List(context.Background(), hcloud.ServerListOpts{
		ListOpts: hcloud.ListOpts{
			LabelSelector: opts.LabelSelector,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("error listing servers: %w", err)
	}
	return p.mapServers(servers), nil
}

func (p *HetznerProvider) AllServers() ([]*types.Server, error) {
	servers, err := p.Client.Server.All(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error listing servers: %w", err)
	}

	return p.mapServers(servers), nil
}

func (p *HetznerProvider) DeleteServer(serverName string, force bool) *types.ServerDeleteStatus {
	if serverName == "" {
		return &types.ServerDeleteStatus{
			Error: fmt.Errorf("empty server name provided"),
		}
	}
	server, _, err := p.Client.Server.Get(context.Background(), serverName)
	if err != nil {
		return &types.ServerDeleteStatus{
			Error: fmt.Errorf("error getting server: %w", err),
		}
	}
	if server == nil {
		p.logger.Debug("Server not found, skipping",
			"server", serverName)
		return &types.ServerDeleteStatus{
			Deleted: true,
		}
	}
	if !force {
		if deleteAfterStr, ok := server.Labels["delete_after"]; ok {
			p.logger.Debug("checking delete_after",
				"server", serverName,
				"delete_after", deleteAfterStr)
			deleteAfter, err := time.Parse("2006-01-02-15-04", deleteAfterStr)
			if err != nil {
				return &types.ServerDeleteStatus{
					Error: fmt.Errorf("invalid delete_after value: %s", deleteAfterStr),
				}
			}
			if err == nil && time.Now().UTC().Before(deleteAfter) {
				p.logger.Warn("server not ready for deletion",
					"server", serverName,
					"delete_after", deleteAfter.Format("2006-01-02 15:04:05"))
				return &types.ServerDeleteStatus{
					DeleteAfter: deleteAfter,
				}
			}
		}
	}
	for _, volume := range server.Volumes {
		_, _, err = p.Client.Volume.Detach(context.Background(), volume)
		if err != nil {
			return &types.ServerDeleteStatus{
				Error: fmt.Errorf("error detaching volume: %w", err),
			}
		}
	}
	_, _, err = p.Client.Server.DeleteWithResult(context.Background(), server)
	if err != nil {
		return &types.ServerDeleteStatus{
			Error: fmt.Errorf("error deleting server: %w", err),
		}
	}
	return &types.ServerDeleteStatus{
		Deleted: true,
	}
}

func (p *HetznerProvider) ServerToCreateOpts(server *types.Server) (options.ServerCreateOpts, error) {
	sshKeys, err := p.KeyNamesToSSHKeys(server.Spec.SSHKeyNames, options.SSHKeyCreateOpts{
		Labels: server.ObjectMeta.Labels,
	})
	if err != nil {
		return options.ServerCreateOpts{}, fmt.Errorf("error converting SSH keys: %w", err)
	}
	cloudInitUserData := fmt.Sprintf(config.DefaultCloudInitUserData, sshKeys[0].Spec.PublicKey)
	return options.ServerCreateOpts{
		Name:     server.ObjectMeta.Name,
		Type:     server.Spec.ServerType,
		Image:    server.Spec.Image,
		Location: server.Spec.Location,
		Provider: "hetzner",
		SSHKeys:  sshKeys,
		UserData: cloudInitUserData,
	}, nil
}

// mapServer converts a Hetzner-specific server to our generic Server type
func (p *HetznerProvider) mapServer(s *hcloud.Server) *types.Server {
	if s == nil {
		return nil
	}

	volumes := make([]*hcloud.Volume, 0)
	for _, volume := range s.Volumes {
		v, _, err := p.Client.Volume.Get(context.Background(), fmt.Sprintf("%d", volume.ID))
		if err != nil {
			p.logger.Error("error getting volume",
				"volume", volume.ID,
				"error", err)
			continue
		}
		volumes = append(volumes, v)
	}
	return &types.Server{
		TypeMeta: types.TypeMeta{
			APIVersion: "v1",
			Kind:       "Server",
		},
		ObjectMeta: types.ObjectMeta{
			Name:   s.Name,
			Labels: s.Labels,
		},
		Spec: types.ServerSpec{
			ServerType: s.ServerType.Name,
			Location:   s.Datacenter.Location.Name,
			Provider:   "hetzner",
			Image:      s.Image.Name,
			Labels:     s.Labels,
			Volumes:    p.mapVolumes(volumes),
			TTL:        s.Labels["ttl"],
		},
		Status: types.ServerStatus{
			Status:      string(s.Status),
			Owner:       s.Labels["owner"],
			Cores:       s.ServerType.Cores,
			Memory:      s.ServerType.Memory,
			Disk:        s.ServerType.Disk,
			PublicNet:   mapPublicNet(&s.PublicNet),
			Created:     s.Created,
			DeleteAfter: timeutil.ParseDeleteAfter(s.Labels["delete_after"]),
		},
	}
}

// mapServers converts a slice of Hetzner servers
func (p *HetznerProvider) mapServers(servers []*hcloud.Server) []*types.Server {
	if servers == nil {
		return nil
	}

	result := make([]*types.Server, len(servers))
	for i, s := range servers {
		result[i] = p.mapServer(s)
	}
	return result
}

func mapPublicNet(publicNet *hcloud.ServerPublicNet) *types.PublicNet {
	if publicNet == nil {
		return nil
	}
	ipv4 := publicNet.IPv4.IP.String()
	return &types.PublicNet{
		IPv4: &struct {
			IP string `json:"ip"`
		}{IP: ipv4},
	}
}
