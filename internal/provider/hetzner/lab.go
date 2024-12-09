package hetzner

import (
	"encoding/json"

	"github.com/pavelanni/storctl/internal/provider/options"
	"github.com/pavelanni/storctl/internal/types"
	"go.etcd.io/bbolt"
)

func (p *HetznerProvider) CreateLab(name string, template string) error {
	return nil
}

func (p *HetznerProvider) GetLabFromCloud(labName string) (*types.Lab, error) {
	lab := &types.Lab{
		TypeMeta: types.TypeMeta{
			APIVersion: "v1",
			Kind:       "Lab",
		},
		ObjectMeta: types.ObjectMeta{
			Name: labName,
		},
	}

	servers, err := p.ListServers(options.ServerListOpts{
		ListOpts: options.ListOpts{
			LabelSelector: "lab_name=" + labName,
		},
	})
	if err != nil {
		return nil, err
	}
	volumes, err := p.ListVolumes(options.VolumeListOpts{
		ListOpts: options.ListOpts{
			LabelSelector: "lab_name=" + labName,
		},
	})
	if err != nil {
		return nil, err
	}
	lab.Status.Servers = append(lab.Status.Servers, servers...)
	lab.Status.Volumes = append(lab.Status.Volumes, volumes...)
	// Add labels from the first server
	if len(servers) > 0 {
		lab.ObjectMeta.Labels = servers[0].ObjectMeta.Labels
	}
	lab.Status.Status = servers[0].Status.Status
	lab.Status.Owner = servers[0].Status.Owner
	lab.Status.Created = servers[0].Status.Created
	lab.Status.DeleteAfter = servers[0].Status.DeleteAfter
	lab.Spec.Location = servers[0].Spec.Location
	lab.Spec.Provider = servers[0].Spec.Provider
	return lab, nil
}

func (p *HetznerProvider) GetLab(labName string) (*types.Lab, error) {
	var lab *types.Lab

	err := p.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(p.labBucket)
		data := b.Get([]byte(labName))
		if data == nil {
			// If not in cache, fetch from cloud
			var err error
			lab, err = p.GetLabFromCloud(labName)
			if err == nil {
				// store in cache
				data, err := json.Marshal(lab)
				if err != nil {
					return err
				}
				return b.Put([]byte(labName), data)
			}
			return err
		}

		lab = &types.Lab{}
		if err := json.Unmarshal(data, lab); err != nil {
			return err
		}
		return nil
	})

	return lab, err
}

func (p *HetznerProvider) SyncLabs() error {
	p.logger.Debug("syncing labs")
	labsMap := make(map[string]*types.Lab)
	allServers, err := p.AllServers()
	if err != nil {
		return err
	}
	// collect unique lab names
	for _, server := range allServers {
		if server.Labels["lab_name"] != "" {
			labsMap[server.Labels["lab_name"]] = &types.Lab{}
		}
	}
	for labName := range labsMap {
		lab, err := p.GetLabFromCloud(labName)
		if err != nil {
			return err
		}
		labsMap[labName] = lab
	}
	return p.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(p.labBucket)

		// Clear existing data
		if err := b.ForEach(func(k, v []byte) error {
			return b.Delete(k)
		}); err != nil {
			return err
		}

		// Store new data
		for labName, lab := range labsMap {
			data, err := json.Marshal(lab)
			if err != nil {
				return err
			}
			if err := b.Put([]byte(labName), data); err != nil {
				return err
			}
		}
		return nil
	})
}

func (p *HetznerProvider) ListLabs(opts options.LabListOpts) ([]*types.Lab, error) {
	var labs []*types.Lab

	err := p.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(p.labBucket)
		return b.ForEach(func(k, v []byte) error {
			var lab types.Lab
			if err := json.Unmarshal(v, &lab); err != nil {
				return err
			}
			labs = append(labs, &lab)
			return nil
		})
	})

	return labs, err
}

func (p *HetznerProvider) DeleteLab(labName string, force bool) *types.LabDeleteStatus {
	lab, err := p.GetLab(labName)
	if err != nil {
		p.logger.Error("failed to get lab details",
			"lab", labName,
			"error", err)
		return &types.LabDeleteStatus{
			Error: err,
		}
	}

	// Get all SSH keys
	sshKeys, err := p.AllSSHKeys()
	if err != nil {
		p.logger.Error("failed to get SSH keys",
			"lab", labName,
			"error", err)
		return &types.LabDeleteStatus{
			Error: err,
		}
	}

	p.logger.Debug("deleting lab",
		"lab", lab.Name,
		"servers", len(lab.Status.Servers),
		"volumes", len(lab.Status.Volumes))

	// Delete volumes first
	for _, volume := range lab.Status.Volumes {
		p.logger.Debug("deleting volume",
			"volume", volume.Name)
		if status := p.DeleteVolume(volume.Name, force); status.Error != nil {
			p.logger.Error("failed to delete volume",
				"volume", volume.Name,
				"error", status.Error)
			return &types.LabDeleteStatus{
				Error: status.Error,
			}
		}
	}

	// Delete servers
	for _, server := range lab.Status.Servers {
		p.logger.Debug("deleting server",
			"server", server.Name)
		if status := p.DeleteServer(server.Name, force); status.Error != nil {
			p.logger.Error("failed to delete server",
				"server", server.Name,
				"error", status.Error)
			return &types.LabDeleteStatus{
				Error: status.Error,
			}
		}
	}

	// Delete SSH keys associated with this lab
	for _, sshKey := range sshKeys {
		if sshKey.Labels["lab_name"] == labName {
			p.logger.Debug("deleting SSH key",
				"key", sshKey.Name)
			if status := p.DeleteSSHKey(sshKey.Name, force); status.Error != nil {
				p.logger.Error("failed to delete SSH key",
					"key", sshKey.Name,
					"error", status.Error)
				return &types.LabDeleteStatus{
					Error: status.Error,
				}
			}
		}
	}

	p.logger.Debug("lab deletion from the cloud completed successfully",
		"lab", labName)

	// Delete from the database
	if err := p.db.Update(func(tx *bbolt.Tx) error {
		return tx.Bucket(p.labBucket).Delete([]byte(labName))
	}); err != nil {
		p.logger.Error("failed to delete lab from the database",
			"lab", labName,
			"error", err)
		return &types.LabDeleteStatus{
			Error: err,
		}
	}

	return &types.LabDeleteStatus{
		Deleted: true,
	}
}
