package lima

// These types are to create a Lima config YAML file
type LimaConfig struct {
	VMType          string                 `yaml:"vmType,omitempty"`
	CPUs            int                    `yaml:"cpus,omitempty"`
	Memory          string                 `yaml:"memory,omitempty"`
	Disk            string                 `yaml:"disk,omitempty"`
	Arch            string                 `yaml:"arch,omitempty"`
	OS              string                 `yaml:"os"`
	Images          []ConfigImage          `yaml:"images"`
	AdditionalDisks []ConfigAdditionalDisk `yaml:"additionalDisks,omitempty"`
	Networks        []ConfigNetwork        `yaml:"networks,omitempty"`
}

type ConfigServer struct {
	Name            string                 `yaml:"name"`
	Role            string                 `yaml:"role"`
	CPUs            int                    `yaml:"cpus"`
	Memory          string                 `yaml:"memory"`
	Disk            string                 `yaml:"disk"`
	Image           string                 `yaml:"image"`
	Arch            string                 `yaml:"arch,omitempty"`
	AdditionalDisks []ConfigAdditionalDisk `yaml:"additionalDisks,omitempty"`
}

type ConfigDisk struct {
	Name     string `yaml:"name"`
	Instance string `yaml:"instance"`
	Size     int    `yaml:"size"` // in bytes
}

type ConfigVolume struct {
	Name   string `yaml:"name"`
	Server string `yaml:"server"`
	Size   string `yaml:"size"`
}

type ConfigAdditionalDisk struct {
	Name   string `yaml:"name"`
	Format bool   `yaml:"format"`
	FsType string `yaml:"fsType,omitempty"`
}

type ConfigNetwork struct {
	LimaNetwork string `yaml:"lima"`
}

type ConfigImage struct {
	Location string `yaml:"location"`
	Arch     string `yaml:"arch"`
	Digest   string `yaml:"digest,omitempty"`
}

type Instance struct {
	Name          string    `json:"name"`
	Hostname      string    `json:"hostname"`
	Status        string    `json:"status"`
	Dir           string    `json:"dir"`
	VMType        string    `json:"vmType"`
	Arch          string    `json:"arch"`
	CPUType       string    `json:"cpuType"`
	CPUs          int       `json:"cpus"`
	Memory        int64     `json:"memory"`
	Disk          int64     `json:"disk"`
	Network       []Network `json:"network"`
	SSHLocalPort  int       `json:"sshLocalPort"`
	SSHConfigFile string    `json:"sshConfigFile"`
	HostAgentPID  int       `json:"hostAgentPID"`
	DriverPID     int       `json:"driverPID"`
	Config        Config    `json:"config"`
	SSHAddress    string    `json:"sshAddress"`
	Protected     bool      `json:"protected"`
	LimaVersion   string    `json:"limaVersion"`
	HostOS        string    `json:"HostOS"`
	HostArch      string    `json:"HostArch"`
	LimaHome      string    `json:"LimaHome"`
	IdentityFile  string    `json:"IdentityFile"`
}

type Network struct {
	Lima       string `json:"lima"`
	MacAddress string `json:"macAddress"`
	Interface  string `json:"interface"`
	Metric     int    `json:"metric"`
}

// These types are to Unmarshal server status we get from limactl list --json

type Config struct {
	VMType               string            `json:"vmType"`
	VMOpts               VMOpts            `json:"vmOpts"`
	OS                   string            `json:"os"`
	Arch                 string            `json:"arch"`
	Images               []Image           `json:"images"`
	CPUType              map[string]string `json:"cpuType"`
	CPUs                 int               `json:"cpus"`
	Memory               string            `json:"memory"`
	Disk                 string            `json:"disk"`
	MountType            string            `json:"mountType"`
	MountInotify         bool              `json:"mountInotify"`
	SSH                  SSH               `json:"ssh"`
	Firmware             Firmware          `json:"firmware"`
	Audio                Audio             `json:"audio"`
	Video                Video             `json:"video"`
	UpgradePackages      bool              `json:"upgradePackages"`
	Containerd           Containerd        `json:"containerd"`
	GuestInstallPrefix   string            `json:"guestInstallPrefix"`
	Networks             []Network         `json:"networks"`
	HostResolver         HostResolver      `json:"hostResolver"`
	PropagateProxyEnv    bool              `json:"propagateProxyEnv"`
	CACerts              CACerts           `json:"caCerts"`
	Rosetta              Rosetta           `json:"rosetta"`
	Plain                bool              `json:"plain"`
	Timezone             string            `json:"timezone"`
	NestedVirtualization bool              `json:"nestedVirtualization"`
	User                 User              `json:"user"`
}

type VMOpts struct {
	Qemu map[string]interface{} `json:"qemu"`
}

type Image struct {
	Location string `json:"location"`
	Arch     string `json:"arch"`
}

type SSH struct {
	LocalPort         int  `json:"localPort"`
	LoadDotSSHPubKeys bool `json:"loadDotSSHPubKeys"`
	ForwardAgent      bool `json:"forwardAgent"`
	ForwardX11        bool `json:"forwardX11"`
	ForwardX11Trusted bool `json:"forwardX11Trusted"`
}

type Firmware struct {
	LegacyBIOS bool `json:"legacyBIOS"`
}

type Audio struct {
	Device string `json:"device"`
}

type Video struct {
	Display string   `json:"display"`
	VNC     VideoVNC `json:"vnc"`
}

type VideoVNC struct {
	Display string `json:"display"`
}

type Containerd struct {
	System   bool      `json:"system"`
	User     bool      `json:"user"`
	Archives []Archive `json:"archives"`
}

type Archive struct {
	Location string `json:"location"`
	Arch     string `json:"arch"`
	Digest   string `json:"digest"`
}

type HostResolver struct {
	Enabled bool `json:"enabled"`
	IPv6    bool `json:"ipv6"`
}

type CACerts struct {
	RemoveDefaults bool `json:"removeDefaults"`
}

type Rosetta struct {
	Enabled bool `json:"enabled"`
	Binfmt  bool `json:"binfmt"`
}

type User struct {
	Name    string `json:"name"`
	Comment string `json:"comment"`
	Home    string `json:"home"`
	UID     int    `json:"uid"`
}
