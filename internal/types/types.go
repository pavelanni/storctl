package types

import (
	"time"
)

// Server represents a generic server across providers
type Server struct {
	ID          string
	Name        string
	Status      string
	Type        string
	Owner       string
	Cores       int
	Memory      float32
	Disk        int
	Location    string
	Labels      map[string]string
	Volumes     []*Volume
	Created     time.Time
	DeleteAfter time.Time
}

type Volume struct {
	ID          string
	Name        string
	Status      string
	Owner       string
	ServerID    string
	ServerName  string
	Location    string
	Size        int
	Format      string
	Labels      map[string]string
	Created     time.Time
	DeleteAfter time.Time
}

type Lab struct {
	ID          string
	Name        string
	Owner       string
	Servers     []*Server
	Volumes     []*Volume
	SSHKeys     []*SSHKey
	DeleteAfter time.Time
}

type SSHKey struct {
	ID          string
	Name        string
	Fingerprint string
	PublicKey   string
	Labels      map[string]string
	Created     time.Time
	DeleteAfter time.Time
}

// Resource represents the common fields for all resources
type Resource struct {
	APIVersion string                 `yaml:"apiVersion"`
	Kind       string                 `yaml:"kind"`
	Metadata   map[string]interface{} `yaml:"metadata"`
	Spec       map[string]interface{} `yaml:"spec"`
}
