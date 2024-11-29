package options

import "github.com/pavelanni/labshop/internal/types"

type ServerCreateOpts struct {
	Name     string
	Type     string          // Server type in cloud provider
	Image    string          // OS image
	Location string          // Datacenter/location
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
