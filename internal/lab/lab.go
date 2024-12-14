package lab

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/pavelanni/storctl/internal/config"
	"github.com/pavelanni/storctl/internal/logger"
	"github.com/pavelanni/storctl/internal/provider"
	"github.com/pavelanni/storctl/internal/provider/options"
	"github.com/pavelanni/storctl/internal/ssh"
	"github.com/pavelanni/storctl/internal/types"
	"github.com/pavelanni/storctl/internal/util/serverchecker"
	"go.etcd.io/bbolt"
)

type Manager interface {
	Create(lab *types.Lab) error
	Get(labName string) (*types.Lab, error)
	List() ([]*types.Lab, error)
	Delete(labName string, force bool) error
	SyncLabs() error
}

type ManagerSvc struct {
	provider   provider.CloudProvider
	sshManager *ssh.Manager
	storage    *Storage
	logger     *slog.Logger
}

type Storage struct {
	db        *bbolt.DB
	labBucket []byte
}

var DefaultManager *ManagerSvc

var _ Manager = (*ManagerSvc)(nil)

func NewBboltDB(path string) (*bbolt.DB, error) {
	db, err := bbolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func NewLabStorage(cfg *config.Config) (*Storage, error) {
	db, err := NewBboltDB(cfg.Storage.Path)
	if err != nil {
		return nil, err
	}

	// Create bucket if it doesn't exist
	err = db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(cfg.Storage.Bucket))
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("create bucket: %w", err)
	}

	return &Storage{
		db:        db,
		labBucket: []byte(cfg.Storage.Bucket),
	}, nil
}

func NewManager(provider provider.CloudProvider, cfg *config.Config) (Manager, error) {
	sshManager := ssh.NewManager(cfg)
	logLevel := logger.ParseLevel(cfg.LogLevel)
	fmt.Println("logLevel", logLevel)
	logger := logger.NewLogger(logLevel)
	storage, err := NewLabStorage(cfg)
	if err != nil {
		return nil, err
	}
	return &ManagerSvc{
		storage:    storage,
		provider:   provider,
		sshManager: sshManager,
		logger:     logger,
	}, nil
}

func (m *ManagerSvc) Create(lab *types.Lab) error {
	labAdminKeyName := strings.Join([]string{lab.ObjectMeta.Name, "admin"}, "-")
	sshKeys := make([]*types.SSHKey, 2) // 2 keys: default admin key and lab admin key
	sshKeys[0] = &types.SSHKey{         // default admin key is already on the cloud
		ObjectMeta: types.ObjectMeta{
			Name: config.DefaultAdminKeyName,
		},
	}
	labAdminPublicKey, err := m.sshManager.CreateLocalKeyPair(labAdminKeyName)
	if err != nil {
		return err
	}
	labAdminCloudKey, err := m.provider.CreateSSHKey(options.SSHKeyCreateOpts{
		Name:      labAdminKeyName,
		PublicKey: labAdminPublicKey,
	})
	if err != nil {
		return err
	}
	sshKeys[1] = labAdminCloudKey

	ttl := lab.Spec.TTL
	if ttl == "" {
		ttl = config.DefaultTTL
	}
	// Create servers
	specServers := lab.Spec.Servers
	servers := make([]*types.Server, 0)
	for _, serverSpec := range specServers {
		s := &types.Server{
			TypeMeta: types.TypeMeta{
				Kind:       "Server",
				APIVersion: "v1",
			},
			ObjectMeta: types.ObjectMeta{
				Name:   strings.Join([]string{lab.ObjectMeta.Name, serverSpec.Name}, "-"),
				Labels: lab.ObjectMeta.Labels,
			},
			Spec: types.ServerSpec{
				Location:   lab.Spec.Location,
				Provider:   lab.Spec.Provider,
				ServerType: serverSpec.ServerType,
				TTL:        ttl,
				Image:      serverSpec.Image,
			},
		}
		result, err := m.provider.CreateServer(options.ServerCreateOpts{
			Name:     s.ObjectMeta.Name,
			Type:     s.Spec.ServerType,
			Image:    s.Spec.Image,
			Location: s.Spec.Location,
			Provider: s.Spec.Provider,
			SSHKeys:  sshKeys,
		})
		if err != nil {
			return err
		}
		servers = append(servers, result)
	}

	// Wait for servers to be ready
	timeout := 30 * time.Minute
	attempts := 20
	results, err := serverchecker.CheckServers(servers, m.logger, timeout, attempts)
	if err != nil {
		return err
	}
	for _, result := range results {
		fmt.Printf("Server %s: Ready: %v\n", result.Server.ObjectMeta.Name, result.Ready)
		if !result.Ready {
			return fmt.Errorf("server %s not ready", result.Server.ObjectMeta.Name)
		}
	}

	// Create volumes
	volumes := lab.Spec.Volumes
	for _, volumeSpec := range volumes {
		if !volumeSpec.Automount { // if not specified, default to false
			volumeSpec.Automount = config.DefaultVolumeAutomount
		}
		if volumeSpec.Format == "" { // if not specified, default to xfs
			volumeSpec.Format = config.DefaultVolumeFormat
		}
		v := &types.Volume{
			TypeMeta: types.TypeMeta{
				Kind:       "Volume",
				APIVersion: "v1",
			},
			ObjectMeta: types.ObjectMeta{
				Name:   strings.Join([]string{lab.ObjectMeta.Name, volumeSpec.Name}, "-"),
				Labels: lab.ObjectMeta.Labels,
			},
			Spec: types.VolumeSpec{
				Size:       volumeSpec.Size,
				ServerName: strings.Join([]string{lab.ObjectMeta.Name, volumeSpec.Server}, "-"),
				Automount:  volumeSpec.Automount,
				Format:     volumeSpec.Format,
			},
		}
		_, err := m.provider.CreateVolume(options.VolumeCreateOpts{
			Name:       v.ObjectMeta.Name,
			Size:       v.Spec.Size,
			ServerName: v.Spec.ServerName,
			Automount:  v.Spec.Automount,
			Format:     v.Spec.Format,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *ManagerSvc) Get(labName string) (*types.Lab, error) {
	lab, err := m.storage.Get(labName)
	if err == nil {
		return lab, nil
	}
	lab, err = m.syncLabFromCloud(labName)
	if err != nil {
		return nil, err
	}
	return lab, nil
}

func (m *ManagerSvc) List() ([]*types.Lab, error) {
	var labs []*types.Lab

	err := m.storage.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(m.storage.labBucket)
		if b == nil {
			return fmt.Errorf("labs bucket not found in database")
		}

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

func (m *ManagerSvc) SyncLabs() error {
	labsMap := make(map[string]*types.Lab)
	allServers, err := m.provider.AllServers()
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
		lab, err := m.getLabFromCloud(labName)
		if err != nil {
			return err
		}
		labsMap[labName] = lab
	}
	return m.storage.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(m.storage.labBucket)

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

func (m *ManagerSvc) Delete(labName string, force bool) error {
	lab, err := m.Get(labName)
	if err != nil {
		return err
	}
	// delete volumes first
	for _, volume := range lab.Status.Volumes {
		m.logger.Debug("deleting volume", "volume", volume.ObjectMeta.Name)
		status := m.provider.DeleteVolume(volume.ObjectMeta.Name, force)
		if status.Error != nil {
			m.logger.Error("failed to delete volume", "volume", volume.ObjectMeta.Name, "error", status.Error)
		}
	}
	// delete servers
	for _, server := range lab.Status.Servers {
		// delete server's ssh keys
		for _, sshKeyName := range server.Spec.SSHKeyNames {
			m.logger.Debug("deleting ssh key", "key", sshKeyName)
			status := m.provider.DeleteSSHKey(sshKeyName, force)
			if status.Error != nil {
				m.logger.Error("failed to delete ssh key", "key", sshKeyName, "error", status.Error)
			}
		}
		m.logger.Debug("deleting server", "server", server.ObjectMeta.Name)
		status := m.provider.DeleteServer(server.ObjectMeta.Name, force)
		if status.Error != nil {
			m.logger.Error("failed to delete server", "server", server.ObjectMeta.Name, "error", status.Error)
		}
	}

	// delete lab from storage
	return m.storage.db.Update(func(tx *bbolt.Tx) error {
		return tx.Bucket(m.storage.labBucket).Delete([]byte(labName))
	})
}

func (s *Storage) Get(labName string) (*types.Lab, error) {
	var lab *types.Lab

	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(s.labBucket)
		data := b.Get([]byte(labName))
		if data == nil {
			return fmt.Errorf("lab %s not found", labName)
		}

		lab = &types.Lab{}
		if err := json.Unmarshal(data, lab); err != nil {
			return err
		}
		return nil
	})

	return lab, err
}

func (s *Storage) Save(lab *types.Lab) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(s.labBucket)
		data, err := json.Marshal(lab)
		if err != nil {
			return err
		}
		return b.Put([]byte(lab.Name), data)
	})
}

func (m *ManagerSvc) syncLabFromCloud(labName string) (*types.Lab, error) {
	lab, err := m.getLabFromCloud(labName)
	if err != nil {
		return nil, err
	}
	if err := m.storage.Save(lab); err != nil {
		m.logger.Warn("failed to save lab to storage", "error", err)
	}
	return lab, nil
}

func (m *ManagerSvc) getLabFromCloud(labName string) (*types.Lab, error) {
	lab := &types.Lab{
		TypeMeta: types.TypeMeta{
			APIVersion: "v1",
			Kind:       "Lab",
		},
		ObjectMeta: types.ObjectMeta{
			Name: labName,
		},
	}

	servers, err := m.provider.ListServers(options.ServerListOpts{
		ListOpts: options.ListOpts{
			LabelSelector: "lab_name=" + labName,
		},
	})
	if err != nil {
		return nil, err
	}
	volumes, err := m.provider.ListVolumes(options.VolumeListOpts{
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
	lab.Status.State = servers[0].Status.Status
	lab.Status.Owner = servers[0].Status.Owner
	lab.Status.Created = servers[0].Status.Created
	lab.Status.DeleteAfter = servers[0].Status.DeleteAfter
	lab.Spec.Location = servers[0].Spec.Location
	lab.Spec.Provider = servers[0].Spec.Provider
	return lab, nil
}
