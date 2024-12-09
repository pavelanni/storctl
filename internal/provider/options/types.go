package options

import "github.com/pavelanni/storctl/internal/types"

type ServerCreateOpts struct {
	Name     string
	Type     string          // Server type in cloud provider
	Image    string          // OS image
	Location string          // Datacenter/location
	Provider string          // Cloud provider
	SSHKeys  []*types.SSHKey // SSH keys
	Labels   map[string]string
	UserData string // cloud-init user data
}

type VolumeCreateOpts struct {
	Name       string
	Size       int
	Location   string
	ServerName string
	Labels     map[string]string
	Automount  bool
	Format     string
}

type SSHKeyCreateOpts struct {
	Name      string
	PublicKey string
	Labels    map[string]string
}

type ListOpts struct {
	Page          int
	PerPage       int
	LabelSelector string
}

type ServerListOpts struct {
	ListOpts
	Name   string
	Status []types.ServerStatus
	Sort   []string
}

type VolumeListOpts struct {
	ListOpts
	Name   string
	Status []types.VolumeStatus
	Sort   []string
}

type LabListOpts struct {
	ListOpts
	Name string
}

type SSHKeyListOpts struct {
	ListOpts
	Name string
}
