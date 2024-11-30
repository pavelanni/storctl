package hetzner

import (
	"encoding/json"

	"github.com/pavelanni/labshop/internal/provider/options"
	"github.com/pavelanni/labshop/internal/types"
	"go.etcd.io/bbolt"
)

func (p *HetznerProvider) CreateLab(name string, template string) error {
	return nil
}

func (p *HetznerProvider) GetLab(labName string) (*types.Lab, error) {
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

func (p *HetznerProvider) GetLabFromDB(labName string) (*types.Lab, error) {
	var lab *types.Lab

	err := p.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(p.labBucket)
		data := b.Get([]byte(labName))
		if data == nil {
			// If not in cache, fetch from cloud
			var err error
			lab, err = p.GetLab(labName)
			return err
		}

		lab = &types.Lab{}
		return json.Unmarshal(data, lab)
	})

	return lab, err
}

func (p *HetznerProvider) SyncLabs() error {
	p.logger.Info("syncing labs")
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
		lab, err := p.GetLab(labName)
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
		"servers", len(lab.Spec.Servers),
		"volumes", len(lab.Spec.Volumes))

	// Delete volumes first
	for _, volume := range lab.Spec.Volumes {
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
	for _, server := range lab.Spec.Servers {
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
