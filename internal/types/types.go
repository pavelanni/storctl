package types

import (
	"time"
)

type TypeMeta struct {
	APIVersion string `json:"apiVersion,omitempty"`
	Kind       string `json:"kind,omitempty"`
}

type ObjectMeta struct {
	Name   string            `json:"name,omitempty"`
	Labels map[string]string `json:"labels,omitempty"`
}

type Server struct {
	TypeMeta   `json:",inline"`
	ObjectMeta `json:"metadata,omitempty"`
	Spec       ServerSpec   `json:"spec"`
	Status     ServerStatus `json:"status,omitempty"`
}

// Server represents a generic server across providers
type ServerSpec struct {
	ServerType  string            `json:"type"`
	Image       string            `json:"image"`
	Location    string            `json:"location"`
	Provider    string            `json:"provider"`
	Labels      map[string]string `json:"labels"`
	Volumes     []*Volume         `json:"volumes"`
	SSHKeyNames []string          `json:"sshKeyNames"`
	TTL         string            `json:"ttl"`
}

type ServerStatus struct {
	Status      string     `json:"status"`
	Owner       string     `json:"owner"`
	Cores       int        `json:"cores"`
	Memory      float32    `json:"memory"`
	Disk        int        `json:"disk"`
	PublicNet   *PublicNet `json:"publicNet"`
	Created     time.Time  `json:"created"`
	DeleteAfter time.Time  `json:"deleteAfter"`
}

type ServerDeleteStatus struct {
	Deleted     bool      `json:"deleted"`
	DeleteAfter time.Time `json:"deleteAfter"`
	Error       error     `json:"error"`
}

type Volume struct {
	TypeMeta   `json:",inline"`
	ObjectMeta `json:"metadata,omitempty"`
	Spec       VolumeSpec   `json:"spec"`
	Status     VolumeStatus `json:"status,omitempty"`
}

type VolumeSpec struct {
	ServerID   string            `json:"serverID"`
	ServerName string            `json:"serverName"`
	Location   string            `json:"location"`
	Provider   string            `json:"provider"`
	Size       int               `json:"size"`
	Automount  bool              `json:"automount"`
	Format     string            `json:"format"`
	Labels     map[string]string `json:"labels"`
	TTL        string            `json:"ttl"`
}

type VolumeStatus struct {
	Status      string    `json:"status"`
	Owner       string    `json:"owner"`
	Created     time.Time `json:"created"`
	DeleteAfter time.Time `json:"deleteAfter"`
}

type VolumeDeleteStatus struct {
	Deleted     bool      `json:"deleted"`
	DeleteAfter time.Time `json:"deleteAfter"`
	Error       error     `json:"error"`
}

type Lab struct {
	TypeMeta   `json:",inline"`
	ObjectMeta `json:"metadata,omitempty"`
	Spec       LabSpec   `json:"spec"`
	Status     LabStatus `json:"status,omitempty"`
}

type LabSpec struct {
	Servers     []*LabServerSpec `json:"servers"`
	Volumes     []*LabVolumeSpec `json:"volumes"`
	TTL         string           `json:"ttl"`
	Provider    string           `json:"provider"`
	Location    string           `json:"location"`
	Ansible     AnsibleSpec      `json:"ansible"`
	CertManager bool             `json:"certManager"`
	LetsEncrypt string           `json:"letsEncrypt"` // prod or staging
}

type LabStatus struct {
	State       string    `json:"state"`
	Owner       string    `json:"owner"`
	Servers     []*Server `json:"servers"`
	Volumes     []*Volume `json:"volumes"`
	Created     time.Time `json:"created"`
	DeleteAfter time.Time `json:"deleteAfter"`
}

type LabServerSpec struct {
	Name       string `json:"name"`
	Role       string `json:"role"`
	ServerType string `json:"type"`
	Image      string `json:"image"`
}

type LabVolumeSpec struct {
	Name      string `json:"name"`
	Server    string `json:"server"`
	Size      int    `json:"size"`
	Format    string `json:"format"`
	Automount bool   `json:"automount"`
}

type LabDeleteStatus struct {
	Deleted     bool      `json:"deleted"`
	DeleteAfter time.Time `json:"deleteAfter"`
	Error       error     `json:"error"`
}

type SSHKey struct {
	TypeMeta   `json:",inline"`
	ObjectMeta `json:"metadata,omitempty"`
	Spec       SSHKeySpec   `json:"spec"`
	Status     SSHKeyStatus `json:"status,omitempty"`
}

type SSHKeySpec struct {
	PublicKey string            `json:"publicKey"`
	Labels    map[string]string `json:"labels"`
	TTL       string            `json:"ttl"`
}

type SSHKeyStatus struct {
	Status      string    `json:"status"`
	Owner       string    `json:"owner"`
	Created     time.Time `json:"created"`
	DeleteAfter time.Time `json:"deleteAfter"`
}

type SSHKeyDeleteStatus struct {
	Deleted     bool      `json:"deleted"`
	DeleteAfter time.Time `json:"deleteAfter"`
	Error       error     `json:"error"`
}

type SSHKeyExistsStatus struct {
	LocalExists  bool      `json:"localExists"`
	CloudExists  bool      `json:"cloudExists"`
	CloudExpired bool      `json:"cloudExpired"`
	DeleteAfter  time.Time `json:"deleteAfter"`
	Error        error     `json:"error"`
}

type AnsibleSpec struct {
	ConfigFile string `json:"configFile"`
	Inventory  string `json:"inventory"`
	Playbook   string `json:"playbook"`
	User       string `json:"user"`
}

// Resource represents the common fields for all resources
type Resource struct {
	TypeMeta   `json:",inline"`
	ObjectMeta `json:"metadata,omitempty"`
	Spec       map[string]interface{} `json:"spec"`
}

type PublicNet struct {
	IPv4 *struct {
		IP string `json:"ip"`
	} `json:"ipv4"`
	FQDN string `json:"fqdn"`
}

type IPv4 struct {
	IP string `json:"ip"`
}
