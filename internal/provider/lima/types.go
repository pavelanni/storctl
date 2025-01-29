package lima

// LimaConfig is the configuration for a Lima VM
type LimaConfig struct {
	VMType          string           `yaml:"vmType,omitempty"`
	CPUs            int              `yaml:"cpus,omitempty"`
	Memory          string           `yaml:"memory,omitempty"`
	Disk            string           `yaml:"disk,omitempty"`
	Arch            string           `yaml:"arch,omitempty"`
	OS              string           `yaml:"os"`
	Images          []Image          `yaml:"images"`
	AdditionalDisks []AdditionalDisk `yaml:"additionalDisks,omitempty"`
	Networks        []Network        `yaml:"networks,omitempty"`
}

type Server struct {
	Name            string           `yaml:"name"`
	Role            string           `yaml:"role"`
	CPUs            int              `yaml:"cpus"`
	Memory          string           `yaml:"memory"`
	Disk            string           `yaml:"disk"`
	Image           string           `yaml:"image"`
	Arch            string           `yaml:"arch,omitempty"`
	AdditionalDisks []AdditionalDisk `yaml:"additionalDisks,omitempty"`
	Status          ServerStatus     `yaml:"status,omitempty"`
}

type ServerStatus struct {
	Host       string `yaml:"host"`
	IP         string `yaml:"ip"`
	SSHAddress string `yaml:"ssh_address"`
	Port       int    `yaml:"port"`
	SSHArgs    string `yaml:"ssh_args"`
	SSHKeyFile string `yaml:"ssh_key_file"`
}

type Volume struct {
	Name   string `yaml:"name"`
	Server string `yaml:"server"`
	Size   string `yaml:"size"`
}

type AdditionalDisk struct {
	Name   string `yaml:"name"`
	Format bool   `yaml:"format"`
	FsType string `yaml:"fsType,omitempty"`
}

type Network struct {
	LimaNetwork string `yaml:"lima"`
}

type Image struct {
	Location string `yaml:"location"`
	Arch     string `yaml:"arch"`
	Digest   string `yaml:"digest,omitempty"`
}
