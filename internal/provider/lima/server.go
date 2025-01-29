package lima

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pavelanni/storctl/internal/config"
	"github.com/pavelanni/storctl/internal/provider/options"
	"github.com/pavelanni/storctl/internal/types"
	"gopkg.in/yaml.v3"
)

var limactlArgs = []string{"--tty=false"}

var serverTypes = map[string]Server{
	"cx22": {
		CPUs:   2,
		Memory: "4GB",
		Disk:   "40GB",
	},
	"cx32": {
		CPUs:   4,
		Memory: "8GB",
		Disk:   "80GB",
	},
	"cx42": {
		CPUs:   8,
		Memory: "16GB",
		Disk:   "160GB",
	},
	"cpx21": {
		CPUs:   2,
		Memory: "4GB",
		Disk:   "80GB",
	},
	"cpx31": {
		CPUs:   4,
		Memory: "8GB",
		Disk:   "160GB",
	},
	"cpx41": {
		CPUs:   8,
		Memory: "16GB",
		Disk:   "240GB",
	},
}

func (p *LimaProvider) CreateServer(opts options.ServerCreateOpts) (*types.Server, error) {
	// Create a Lima config file in the DefaultLimaDir using the provided opts
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	limaDir := filepath.Join(homeDir, config.DefaultConfigDir, config.DefaultLimaDir)
	limaConfigFile := filepath.Join(limaDir, opts.Name+".yaml")
	server, ok := serverTypes[opts.Type]
	if !ok {
		return nil, fmt.Errorf("invalid server type: %s", opts.Type)
	}
	if opts.AdditionalDisks != nil {
		additionalDisks := []AdditionalDisk{}
		for _, disk := range opts.AdditionalDisks {
			additionalDisks = append(additionalDisks, AdditionalDisk{
				Name:   disk.Name,
				Format: disk.Format,
				FsType: disk.FsType,
			})
		}
		server.AdditionalDisks = additionalDisks
	}
	arch := getArchForArch(p.arch)
	server.Name = opts.Name
	server.Image = opts.Image
	server.Arch = arch
	limaConfig := createLimaConfig(server)
	if err := writeConfig(limaConfigFile, limaConfig); err != nil {
		return nil, fmt.Errorf("error writing config for %s: %v", opts.Name, err)
	}

	if err := createVM(opts.Name, limaConfigFile); err != nil {
		return nil, fmt.Errorf("error creating VM for %s: %v", opts.Name, err)
	}

	return nil, nil
}

func (p *LimaProvider) GetServer(name string) (*types.Server, error) {
	return nil, nil
}

func (p *LimaProvider) ListServers(opts options.ServerListOpts) ([]*types.Server, error) {
	return nil, nil
}

func (p *LimaProvider) AllServers() ([]*types.Server, error) {
	return nil, nil
}

func (p *LimaProvider) DeleteServer(name string, force bool) *types.ServerDeleteStatus {
	return nil
}

func (p *LimaProvider) ServerToCreateOpts(server *types.Server) (options.ServerCreateOpts, error) {
	return options.ServerCreateOpts{}, nil
}

func createLimaConfig(server Server) LimaConfig {
	// Parse Ubuntu version from image string
	version := strings.TrimPrefix(server.Image, "ubuntu-")

	limaConfig := LimaConfig{
		VMType: "qemu",
		CPUs:   server.CPUs,
		Memory: server.Memory,
		Disk:   server.Disk,
		Arch:   getArchForArch(server.Arch),
		OS:     "Linux",
		Images: []Image{
			{
				Location: fmt.Sprintf("https://cloud-images.ubuntu.com/releases/%s/release/ubuntu-%s-server-cloudimg-%s.img",
					version, version, getArchForImage(server.Arch)),
				Arch: getArchForArch(server.Arch),
			},
		},
		Networks: []Network{
			{
				LimaNetwork: "shared",
			},
		},
		AdditionalDisks: server.AdditionalDisks,
	}

	return limaConfig
}

func createVM(name, configPath string) error {
	listCmd := exec.Command("limactl", "list")
	output, err := listCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("listing VMs: %w", err)
	}

	if strings.Contains(string(output), name) {
		fmt.Printf("VM %s already exists\n", name)
		return nil
	}

	createCmd := exec.Command("limactl", append(limactlArgs, "start", "--name", name, configPath)...)
	createCmd.Stdout = os.Stdout
	createCmd.Stderr = os.Stderr

	fmt.Printf("Creating VM %s...\n", name)
	if err := createCmd.Run(); err != nil {
		return fmt.Errorf("creating VM: %w", err)
	}

	fmt.Printf("Successfully created VM %s\n", name)
	return nil
}

func writeConfig(filename string, config LimaConfig) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("writing file: %w", err)
	}

	return nil
}

// Ubuntu images use arm64 and amd64, but Lima uses aarch64 and x86_64
// So we need ArchForImage and ArchForArch (below)
func getArchForImage(arch string) string {
	switch arch {
	case "amd64":
		return "amd64"
	case "x86_64":
		return "amd64"
	case "arm64":
		return "arm64"
	case "aarch64":
		return "arm64"

	default:
		return "arm64" // since we expect it to be running usually on Mac M-series
	}
}

// Lima uses x86_64 and aarch64, but Ubuntu images use amd64 and arm64
// So we need ArchForImage and ArchForArch (above)
func getArchForArch(arch string) string {
	switch arch {
	case "amd64":
		return "x86_64"
	case "x86_64":
		return "x86_64"
	case "arm64":
		return "aarch64"
	case "aarch64":
		return "aarch64"

	default:
		return "aarch64" // since we expect it to be running usually on Mac M-series
	}
}
