package virt

import (
	"github.com/digitalocean/go-libvirt"
)

// These types are to create a libvirt config XML file
type VMConfig struct {
	Name        string
	VCPU        string
	Memory      string
	Disk        string
	Network     string
	VNCListener string
}

type ConfigServer struct {
	Name   string `yaml:"name"`
	Role   string `yaml:"role"`
	CPUs   int    `yaml:"cpus"`
	Memory string `yaml:"memory"`
	Disk   string `yaml:"disk"`
	Image  string `yaml:"image"`
	Arch   string `yaml:"arch,omitempty"`
}

type VolumeConfig struct {
	VolumeName string
	VolumePool libvirt.StoragePool
	VolumePath string
	VolumeSize int64
}

type StoragePoolConfig struct {
	PoolName  string
	PoolPath  string
	PoolMode  string
	PoolOwner string
	PoolGroup string
}

type NetworkConfig struct {
	NetworkName    string
	BridgeName     string
	IPAddress      string
	Netmask        string
	DHCPRangeStart string
	DHCPRangeEnd   string
}
