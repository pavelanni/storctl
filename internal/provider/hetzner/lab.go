package hetzner

import (
	"github.com/pavelanni/labshop/internal/types"
)

func (p *HetznerProvider) CreateLab(name string, template string) error {
	return nil
}

func (p *HetznerProvider) GetLab(labName string) (*types.Lab, error) {
	lab := &types.Lab{Name: labName}

	servers, err := p.AllServers()
	if err != nil {
		return nil, err
	}
	volumes, err := p.AllVolumes()
	if err != nil {
		return nil, err
	}
	for _, server := range servers {
		labName := server.Labels["lab_name"]
		if labName == "" {
			continue
		}
		if labName != lab.Name {
			continue
		}
		lab.Servers = append(lab.Servers, server)
	}
	for _, volume := range volumes {
		labName := volume.Labels["lab_name"]
		if labName != lab.Name {
			continue
		}
		lab.Volumes = append(lab.Volumes, volume)
	}
	return lab, nil
}

func (p *HetznerProvider) DeleteLab(labName string, force bool) error {
	p.logger.Info("starting lab deletion",
		"lab", labName,
		"force", force)

	lab, err := p.GetLab(labName)
	if err != nil {
		p.logger.Error("failed to get lab details",
			"lab", labName,
			"error", err)
		return err
	}

	// Get all SSH keys
	sshKeys, err := p.AllSSHKeys()
	if err != nil {
		p.logger.Error("failed to get SSH keys",
			"lab", labName,
			"error", err)
		return err
	}

	p.logger.Info("deleting lab",
		"lab", lab.Name,
		"servers", len(lab.Servers),
		"volumes", len(lab.Volumes))

	// Delete volumes first
	for _, volume := range lab.Volumes {
		p.logger.Info("deleting volume",
			"volume", volume.Name)
		if err := p.DeleteVolume(volume.Name, force); err != nil {
			p.logger.Error("failed to delete volume",
				"volume", volume.Name,
				"error", err)
			return err
		}
	}

	// Delete servers
	for _, server := range lab.Servers {
		p.logger.Info("deleting server",
			"server", server.Name)
		if err := p.DeleteServer(server.Name, force); err != nil {
			p.logger.Error("failed to delete server",
				"server", server.Name,
				"error", err)
			return err
		}
	}

	// Delete SSH keys associated with this lab
	for _, sshKey := range sshKeys {
		if sshKey.Labels["lab_name"] == labName {
			p.logger.Info("deleting SSH key",
				"key", sshKey.Name)
			if err := p.DeleteSSHKey(sshKey.Name, force); err != nil {
				p.logger.Error("failed to delete SSH key",
					"key", sshKey.Name,
					"error", err)
				return err
			}
		}
	}

	p.logger.Info("lab deletion completed successfully",
		"lab", labName)
	return nil
}
