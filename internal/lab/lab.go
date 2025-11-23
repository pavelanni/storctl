// Package lab contains the lab manager for the storctl tool.
// It includes the functions to create, get, list, and delete labs.
// It also includes the functions to sync labs from the provider and create an ansible inventory file.
// Each lab is stored in the local storage and can be retrieved later.
// The lab manager also includes the functions to create an ansible inventory file and run an ansible playbook.
package lab

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/pavelanni/storctl/internal/config"
	"github.com/pavelanni/storctl/internal/logger"
	"github.com/pavelanni/storctl/internal/provider"
	"github.com/pavelanni/storctl/internal/provider/options"
	"github.com/pavelanni/storctl/internal/ssh"
	"github.com/pavelanni/storctl/internal/storage"
	"github.com/pavelanni/storctl/internal/storage/local"
	"github.com/pavelanni/storctl/internal/storage/postgres"
	"github.com/pavelanni/storctl/internal/types"
	"github.com/pavelanni/storctl/internal/util/serverchecker"
)

type Manager interface {
	Create(lab *types.Lab) error
	Get(labName string) (*types.Lab, error)
	List() ([]*types.Lab, error)
	Delete(labName string, force bool) error
	SyncLabs() error
	CreateAnsibleInventoryFile(lab *types.Lab) error
	RunAnsiblePlaybook(lab *types.Lab) error
}

type ManagerSvc struct {
	Provider   provider.CloudProvider
	SshManager *ssh.Manager
	Storage    storage.Storage
	Logger     *slog.Logger
}

var DefaultManager *ManagerSvc

var _ Manager = (*ManagerSvc)(nil)

func NewManager(provider provider.CloudProvider, cfg *config.Config) (*ManagerSvc, error) {
	sshManager := ssh.NewManager(cfg)
	var storage storage.Storage
	var err error

	switch cfg.Storage.Type {
	case "postgres":
		storage, err = postgres.New(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to create lab storage: %w", err)
		}
	case "local", "": // empty string defaults to local for backward compatibility
		storage, err = local.New(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to create lab storage: %w", err)
		}
	default:
		return nil, fmt.Errorf("invalid storage type: %s", cfg.Storage.Type)
	}
	return &ManagerSvc{Storage: storage, Provider: provider, SshManager: sshManager, Logger: logger.Get()}, nil
}

// Create creates a new lab
// It creates the lab in the cloud and stores the lab in the local storage
// It creates servers, volumes, and ssh keys
func (m *ManagerSvc) Create(lab *types.Lab) error {
	switch lab.Spec.Provider {
	case "lima":
		err := m.createLabLima(lab)
		if err != nil {
			return fmt.Errorf("failed to create lab: %w", err)
		}
	case "hetzner":
		err := m.createLabHetzner(lab)
		if err != nil {
			return fmt.Errorf("failed to create lab: %w", err)
		}
	}
	m.Logger.Debug("created lab", "lab", lab)
	m.Logger.Debug("lab servers:")
	for _, server := range lab.Status.Servers {
		m.Logger.Debug("server", "server", server)
	}
	m.Logger.Debug("lab volumes:")
	for _, volume := range lab.Status.Volumes {
		m.Logger.Debug("volume", "volume", volume)
	}
	err := m.Storage.Save(lab)
	if err != nil {
		return fmt.Errorf("failed to save lab: %w", err)
	}
	return nil
}

func (m *ManagerSvc) Get(labName string) (*types.Lab, error) {
	if m == nil {
		return nil, fmt.Errorf("manager is nil")
	}
	if m.Storage == nil {
		return nil, fmt.Errorf("storage is nil")
	}
	lab, err := m.Storage.Get(labName)
	if err == nil {
		return lab, nil
	}
	lab, err = m.syncLabFromProvider(labName)
	if err != nil {
		return nil, fmt.Errorf("failed to sync lab from provider: %w", err)
	}
	return lab, nil
}

func (m *ManagerSvc) List() ([]*types.Lab, error) {
	var labs []*types.Lab

	labs, err := m.Storage.List()
	if err != nil {
		return nil, fmt.Errorf("failed to list labs: %w", err)
	}
	return labs, nil
}

func (m *ManagerSvc) SyncLabs() error {
	labsMap := make(map[string]*types.Lab)
	allServers, err := m.Provider.AllServers()
	if err != nil {
		return fmt.Errorf("failed to get all servers: %w", err)
	}
	// collect unique lab names
	for _, server := range allServers {
		if server.Labels["lab_name"] != "" {
			labsMap[server.Labels["lab_name"]] = &types.Lab{}
		}
	}
	for labName := range labsMap {
		lab, err := m.getLabFromProvider(labName)
		if err != nil {
			return fmt.Errorf("failed to get lab from provider: %w", err)
		}
		labsMap[labName] = lab
	}
	for _, lab := range labsMap {
		err := m.Storage.Save(lab)
		if err != nil {
			return fmt.Errorf("failed to save lab: %w", err)
		}
	}
	return nil
}

func (m *ManagerSvc) Delete(labName string, force bool) error {
	var err error
	switch m.Provider.Name() {
	case "lima":
		err = m.deleteLabLima(labName, force)
	case "hetzner":
		err = m.deleteLabHetzner(labName, force)
	}
	if err != nil {
		return fmt.Errorf("failed to delete lab: %w", err)
	}
	err = m.Storage.Delete(labName)
	if err != nil {
		return fmt.Errorf("failed to delete lab from storage: %w", err)
	}
	return nil
}

func (m *ManagerSvc) syncLabFromProvider(labName string) (*types.Lab, error) {
	lab, err := m.getLabFromProvider(labName)
	if err != nil {
		return nil, fmt.Errorf("failed to get lab from provider: %w", err)
	}
	if err := m.Storage.Save(lab); err != nil {
		return nil, fmt.Errorf("failed to save lab to storage: %w", err)
	}
	return lab, nil
}

func (m *ManagerSvc) getLabFromProvider(labName string) (*types.Lab, error) {
	lab := &types.Lab{
		TypeMeta: types.TypeMeta{
			APIVersion: "v1",
			Kind:       "Lab",
		},
		ObjectMeta: types.ObjectMeta{
			Name: labName,
		},
	}

	servers, err := m.Provider.ListServers(options.ServerListOpts{
		ListOpts: options.ListOpts{
			LabelSelector: "lab_name=" + labName,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list servers: %w", err)
	}
	volumes, err := m.Provider.ListVolumes(options.VolumeListOpts{
		ListOpts: options.ListOpts{
			LabelSelector: "lab_name=" + labName,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list volumes: %w", err)
	}
	lab.Status.Servers = append(lab.Status.Servers, servers...)
	lab.Status.Volumes = append(lab.Status.Volumes, volumes...)
	// Add labels from the first server
	if len(servers) > 0 {
		lab.Labels = servers[0].Labels
		lab.Status.State = servers[0].Status.Status
		lab.Status.Owner = servers[0].Status.Owner
		lab.Status.Created = servers[0].Status.Created
		lab.Status.DeleteAfter = servers[0].Status.DeleteAfter
		lab.Spec.Location = servers[0].Spec.Location
		lab.Spec.Provider = servers[0].Spec.Provider
	}
	return lab, nil
}

func (m *ManagerSvc) createLabLima(lab *types.Lab) error {
	volumes := lab.Spec.Volumes
	volumesStatus := make([]*types.Volume, len(volumes))
	for i, volume := range volumes {
		fmt.Printf("Creating volume %s of size %dGB...\n", strings.Join([]string{lab.Name, volume.Name}, "-"), volume.Size)
		volume, err := m.Provider.CreateVolume(options.VolumeCreateOpts{
			Name:       strings.Join([]string{lab.Name, volume.Name}, "-"),
			Size:       volume.Size,
			ServerName: volume.Server,
			Automount:  volume.Automount,
			Format:     volume.Format,
			Labels:     lab.Labels,
		})
		if err != nil {
			return fmt.Errorf("failed to create volume: %w", err)
		}
		volumesStatus[i] = volume
		m.Logger.Debug("created volume", "volume", volume)
	}
	lab.Status.Volumes = volumesStatus
	m.Logger.Debug("created volumes:")
	for _, volume := range volumesStatus {
		m.Logger.Debug("volume", "volume", volume)
	}
	additionalDisks := make(map[string][]options.AdditionalDisk)
	for _, volume := range volumes {
		labServerName := strings.Join([]string{lab.Name, volume.Server}, "-")
		additionalDisks[labServerName] = append(additionalDisks[labServerName], options.AdditionalDisk{
			Name:   strings.Join([]string{lab.Name, volume.Name}, "-"),
			Format: false,
		})
	}
	specServers := lab.Spec.Servers
	serversStatus := make([]*types.Server, len(specServers))
	for i, serverSpec := range specServers {
		s := &types.Server{
			TypeMeta: types.TypeMeta{
				Kind:       "Server",
				APIVersion: "v1",
			},
			ObjectMeta: types.ObjectMeta{
				Name:   strings.Join([]string{lab.Name, serverSpec.Name}, "-"),
				Labels: lab.Labels,
			},
			Spec: types.ServerSpec{
				Location:   lab.Spec.Location,
				Provider:   lab.Spec.Provider,
				ServerType: serverSpec.ServerType,
				Image:      serverSpec.Image,
			},
		}
		serverAdditionalDisks, ok := additionalDisks[s.Name]
		if !ok {
			serverAdditionalDisks = []options.AdditionalDisk{}
		}
		fmt.Printf("Creating server %s with additional disks: %v\n", s.Name, serverAdditionalDisks)
		server, err := m.Provider.CreateServer(options.ServerCreateOpts{
			Name:            s.Name,
			Type:            s.Spec.ServerType,
			Image:           s.Spec.Image,
			Location:        s.Spec.Location,
			Provider:        s.Spec.Provider,
			Labels:          s.Labels,
			AdditionalDisks: serverAdditionalDisks,
		})
		if err != nil {
			return fmt.Errorf("failed to create server: %w", err)
		}
		m.Logger.Debug("created server", "server", server)
		serversStatus[i] = server
	}
	lab.Status.Servers = serversStatus
	m.Logger.Debug("created servers:")
	for _, server := range serversStatus {
		m.Logger.Debug("server", "server", server)
	}
	return nil
}

func (m *ManagerSvc) createLabHetzner(lab *types.Lab) error {
	labAdminKeyName := strings.Join([]string{lab.Name, "admin"}, "-")
	sshKeys := make([]*types.SSHKey, 2) // 2 keys: default admin key and lab admin key
	sshKeys[0] = &types.SSHKey{         // default admin key is already on the cloud
		TypeMeta: types.TypeMeta{
			Kind:       "SSHKey",
			APIVersion: "v1",
		},
		ObjectMeta: types.ObjectMeta{
			Name:   config.DefaultAdminKeyName,
			Labels: lab.Labels,
		},
	}
	fmt.Printf("Creating lab admin key %s...\n", labAdminKeyName)
	labAdminPublicKey, err := m.SshManager.CreateLocalKeyPair(labAdminKeyName)
	if err != nil {
		return fmt.Errorf("failed to create lab admin key: %w", err)
	}
	labAdminCloudKey, err := m.Provider.CreateSSHKey(options.SSHKeyCreateOpts{
		Name:      labAdminKeyName,
		PublicKey: labAdminPublicKey,
	})
	if err != nil {
		return fmt.Errorf("failed to create lab admin cloud key: %w", err)
	}
	sshKeys[1] = labAdminCloudKey

	ttl := lab.Spec.TTL
	if ttl == "" {
		ttl = config.DefaultTTL
	}
	// Create servers
	serversString := ""
	for _, serverSpec := range lab.Spec.Servers {
		serversString += serverSpec.Name + ", "
	}
	fmt.Printf("Creating %d servers: %s\n", len(lab.Spec.Servers), serversString)
	specServers := lab.Spec.Servers
	servers := make([]*types.Server, 0)
	for _, serverSpec := range specServers {
		s := &types.Server{
			TypeMeta: types.TypeMeta{
				Kind:       "Server",
				APIVersion: "v1",
			},
			ObjectMeta: types.ObjectMeta{
				Name:   strings.Join([]string{lab.Name, serverSpec.Name}, "-"),
				Labels: lab.Labels,
			},
			Spec: types.ServerSpec{
				Location:   lab.Spec.Location,
				Provider:   lab.Spec.Provider,
				ServerType: serverSpec.ServerType,
				TTL:        ttl,
				Image:      serverSpec.Image,
			},
		}
		fmt.Printf("Creating server %s...\n", s.Name)
		result, err := m.Provider.CreateServer(options.ServerCreateOpts{
			Name:     s.Name,
			Type:     s.Spec.ServerType,
			Image:    s.Spec.Image,
			Location: s.Spec.Location,
			Provider: s.Spec.Provider,
			SSHKeys:  sshKeys,
			Labels:   s.Labels,
			UserData: fmt.Sprintf(config.DefaultCloudInitUserData, labAdminPublicKey),
		})
		if err != nil {
			return fmt.Errorf("failed to create server: %w", err)
		}
		servers = append(servers, result)
	}

	// Wait for servers to be ready
	fmt.Println("Waiting for servers to be ready...")
	timeout := 30 * time.Minute
	attempts := 20
	results, err := serverchecker.CheckServers(servers, m.Logger, timeout, attempts)
	if err != nil {
		return fmt.Errorf("failed to check servers: %w", err)
	}
	for _, result := range results {
		fmt.Printf("Server %s: Ready: %v\n", result.Server.Name, result.Ready)
		if !result.Ready {
			return fmt.Errorf("server %s not ready", result.Server.Name)
		}
	}
	fmt.Println("Servers are ready")
	// Create volumes
	volumesString := ""
	for _, volumeSpec := range lab.Spec.Volumes {
		volumesString += volumeSpec.Name + ", "
	}
	fmt.Printf("Creating %d volumes: %s\n", len(lab.Spec.Volumes), volumesString)
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
				Name:   strings.Join([]string{lab.Name, volumeSpec.Name}, "-"),
				Labels: lab.Labels,
			},
			Spec: types.VolumeSpec{
				Size:       volumeSpec.Size,
				ServerName: strings.Join([]string{lab.Name, volumeSpec.Server}, "-"),
				Automount:  volumeSpec.Automount,
				Format:     volumeSpec.Format,
			},
		}
		fmt.Printf("Creating volume %s...\n", v.Name)
		_, err := m.Provider.CreateVolume(options.VolumeCreateOpts{
			Name:       v.Name,
			Size:       v.Spec.Size,
			ServerName: v.Spec.ServerName,
			Automount:  v.Spec.Automount,
			Format:     v.Spec.Format,
			Labels:     v.Labels,
		})
		if err != nil {
			return fmt.Errorf("failed to create volume: %w", err)
		}
	}

	return nil
}

func (m *ManagerSvc) deleteLabLima(labName string, force bool) error {
	lab, err := m.Get(labName)
	if err != nil {
		return fmt.Errorf("failed to get lab: %w", err)
	}
	// in Lima, delete servers first
	for _, server := range lab.Status.Servers {
		// delete server's ssh keys
		for _, sshKeyName := range server.Spec.SSHKeyNames {
			m.Logger.Info("deleting ssh key", "key", sshKeyName)
			status := m.Provider.DeleteSSHKey(sshKeyName, force)
			if status.Error != nil {
				return fmt.Errorf("failed to delete ssh key %s: %w", sshKeyName, status.Error)
			}
		}
		m.Logger.Info("deleting server", "server", server.Name)
		status := m.Provider.DeleteServer(server.Name, force)
		if status.Error != nil {
			return fmt.Errorf("failed to delete server %s: %w", server.Name, status.Error)
		}
	}

	// delete volumes after servers
	for _, volume := range lab.Status.Volumes {
		m.Logger.Info("deleting volume", "volume", volume.Name)
		status := m.Provider.DeleteVolume(volume.Name, force)
		if status.Error != nil {
			return fmt.Errorf("failed to delete volume %s: %w", volume.Name, status.Error)
		}
	}
	return nil
}

func (m *ManagerSvc) deleteLabHetzner(labName string, force bool) error {
	lab, err := m.Get(labName)
	if err != nil {
		return fmt.Errorf("failed to get lab: %w", err)
	}
	// Check if the lab is ready for deletion
	if !lab.Status.DeleteAfter.Before(time.Now().UTC()) && !force {
		return fmt.Errorf("lab %s is not ready for deletion", labName)
	}
	// delete volumes first
	for _, volume := range lab.Status.Volumes {
		m.Logger.Info("deleting volume", "volume", volume.Name)
		status := m.Provider.DeleteVolume(volume.Name, force)
		if status.Error != nil {
			return fmt.Errorf("failed to delete volume %s: %w", volume.Name, status.Error)
		}
	}
	// delete servers
	for _, server := range lab.Status.Servers {
		// delete server's ssh keys
		for _, sshKeyName := range server.Spec.SSHKeyNames {
			m.Logger.Info("deleting ssh key", "key", sshKeyName)
			status := m.Provider.DeleteSSHKey(sshKeyName, force)
			if status.Error != nil {
				return fmt.Errorf("failed to delete ssh key %s: %w", sshKeyName, status.Error)
			}
		}
		m.Logger.Info("deleting server", "server", server.Name)
		status := m.Provider.DeleteServer(server.Name, force)
		if status.Error != nil {
			return fmt.Errorf("failed to delete server %s: %w", server.Name, status.Error)
		}
	}
	return nil
}
