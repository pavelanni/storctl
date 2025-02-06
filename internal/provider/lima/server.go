package lima

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/pavelanni/storctl/internal/config"
	"github.com/pavelanni/storctl/internal/provider/options"
	"github.com/pavelanni/storctl/internal/types"
	"gopkg.in/yaml.v3"
)

var limactlArgs = []string{"--tty=false"}

var serverTypes = map[string]ConfigServer{
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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	if opts.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if opts.Type == "" {
		return nil, fmt.Errorf("type is required")
	}
	if opts.Image == "" {
		return nil, fmt.Errorf("image is required")
	}
	checkServer, err := p.GetServer(opts.Name)
	if err != nil {
		if !strings.Contains(err.Error(), "not found") {
			return nil, fmt.Errorf("checking server: %w", err)
		}
	}
	if checkServer != nil {
		fmt.Println("server already exists", checkServer.ObjectMeta.Name)
		return nil, fmt.Errorf("server %s already exists", checkServer.ObjectMeta.Name)
	}
	// Create a Lima config file in the DefaultLimaDir using the provided opts
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("error getting home directory: %w", err)
	}
	limaDir := filepath.Join(homeDir, config.DefaultConfigDir, config.DefaultLimaDir)
	if _, err := os.Stat(limaDir); os.IsNotExist(err) {
		err = os.MkdirAll(limaDir, 0755)
		if err != nil {
			return nil, fmt.Errorf("error creating Lima directory: %w", err)
		}
	}
	limaConfigFile := filepath.Join(limaDir, opts.Name+".yaml")
	server, ok := serverTypes[opts.Type]
	if !ok {
		return nil, fmt.Errorf("invalid server type: %s", opts.Type)
	}
	if len(opts.AdditionalDisks) > 0 {
		fmt.Printf("Additional disks in Lima provider from opts: %v\n", opts.AdditionalDisks)
		additionalDisks := []ConfigAdditionalDisk{}
		for _, disk := range opts.AdditionalDisks {
			additionalDisks = append(additionalDisks, ConfigAdditionalDisk{
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
		return nil, fmt.Errorf("error writing config for %s: %w", opts.Name, err)
	}

	if err := createVM(ctx, opts.Name, limaConfigFile); err != nil {
		return nil, fmt.Errorf("error creating VM for %s: %w", opts.Name, err)
	}

	newServer, err := p.GetServer(opts.Name)
	if err != nil {
		return nil, fmt.Errorf("error getting server for %s: %w", opts.Name, err)
	}
	return newServer, nil
}

func (p *LimaProvider) GetServer(name string) (*types.Server, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	cmd := exec.CommandContext(ctx, "limactl", "list", "--json", name)
	output, err := cmd.Output()
	if err != nil {
		// If there's an error but output is empty, it's likely because the server doesn't exist
		if len(strings.TrimSpace(string(output))) == 0 {
			return nil, nil
		}
		return nil, fmt.Errorf("listing servers: %w", err)
	}

	// Empty output also means no server
	if len(strings.TrimSpace(string(output))) == 0 {
		return nil, nil
	}

	// Parse the output
	var limaServer Instance
	for _, line := range strings.Split(string(output), "\n") {
		if line == "" {
			continue
		}
		if err := json.Unmarshal([]byte(line), &limaServer); err != nil {
			return nil, fmt.Errorf("unmarshalling server: %s: %w", line, err)
		}
		server, err := mapServer(limaServer)
		if err != nil {
			return nil, fmt.Errorf("error mapping server: %w", err)
		}
		return server, nil
	}

	// No valid data found
	return nil, nil
}

func (p *LimaProvider) ListServers(opts options.ServerListOpts) ([]*types.Server, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var labName string

	if opts.ListOpts.LabelSelector != "" {
		label := opts.ListOpts.LabelSelector
		labName = strings.TrimPrefix(label, "lab_name=")
	}
	cmd := exec.CommandContext(ctx, "limactl", "list", "--json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error listing servers: %w", err)
	}
	servers := []*types.Server{}
	for _, line := range strings.Split(string(output), "\n") {
		if line == "" {
			continue
		}
		limaServer := Instance{}
		err := json.Unmarshal([]byte(line), &limaServer)
		if err != nil {
			return nil, fmt.Errorf("error unmarshalling server: %w", err)
		}
		if strings.HasPrefix(limaServer.Name, labName) {
			server, err := mapServer(limaServer)
			if err != nil {
				return nil, fmt.Errorf("error mapping server: %w", err)
			}
			servers = append(servers, server)
		}
	}
	return servers, nil
}

func (p *LimaProvider) AllServers() ([]*types.Server, error) {
	servers, err := p.ListServers(options.ServerListOpts{})
	if err != nil {
		return nil, fmt.Errorf("error listing servers: %w", err)
	}
	return servers, nil
}

func (p *LimaProvider) DeleteServer(name string, force bool) *types.ServerDeleteStatus {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if name == "" {
		return &types.ServerDeleteStatus{
			Deleted: false,
			Error:   fmt.Errorf("name is required"),
		}
	}
	err := exec.CommandContext(ctx, "limactl", "delete", name, "--force").Run()
	if err != nil {
		return &types.ServerDeleteStatus{
			Deleted: false,
			Error:   fmt.Errorf("deleting server: %w", err),
		}
	}
	return &types.ServerDeleteStatus{
		Deleted: true,
	}
}

func (p *LimaProvider) ServerToCreateOpts(server *types.Server) (options.ServerCreateOpts, error) {
	sshKeys, err := p.KeyNamesToSSHKeys(server.Spec.SSHKeyNames, options.SSHKeyCreateOpts{
		Labels: server.ObjectMeta.Labels,
	})
	if err != nil {
		return options.ServerCreateOpts{}, err
	}
	cloudInitUserData := fmt.Sprintf(config.DefaultCloudInitUserData, sshKeys[0].Spec.PublicKey)
	return options.ServerCreateOpts{
		Name:     server.ObjectMeta.Name,
		Type:     server.Spec.ServerType,
		Image:    server.Spec.Image,
		Location: server.Spec.Location,
		Provider: "lima",
		SSHKeys:  sshKeys,
		UserData: cloudInitUserData,
	}, nil
}

func createLimaConfig(server ConfigServer) LimaConfig {
	// Parse Ubuntu version from image string
	version := strings.TrimPrefix(server.Image, "ubuntu-")

	limaConfig := LimaConfig{
		VMType: "qemu",
		CPUs:   server.CPUs,
		Memory: server.Memory,
		Disk:   server.Disk,
		Arch:   getArchForArch(server.Arch),
		OS:     "Linux",
		Images: []ConfigImage{
			{
				Location: fmt.Sprintf("https://cloud-images.ubuntu.com/releases/%s/release/ubuntu-%s-server-cloudimg-%s.img",
					version, version, getArchForImage(server.Arch)),
				Arch: getArchForArch(server.Arch),
			},
		},
		Networks: []ConfigNetwork{
			{
				LimaNetwork: "shared",
			},
		},
		AdditionalDisks: server.AdditionalDisks,
	}

	return limaConfig
}

func createVM(ctx context.Context, name, configPath string) error {
	if name == "" {
		return fmt.Errorf("name is required")
	}
	if configPath == "" {
		return fmt.Errorf("configPath is required")
	}

	createCmd := exec.CommandContext(ctx, "limactl", append(limactlArgs, "start", "--name", name, configPath)...)
	createCmd.Stdout = os.Stdout
	createCmd.Stderr = os.Stderr

	fmt.Printf("Creating VM %s...\n", name)
	if err := createCmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("timeout while creating VM: %w", err)
		}
		return fmt.Errorf("error creating VM: %w", err)
	}

	fmt.Printf("Successfully created VM %s\n", name)
	return nil
}

func writeConfig(filename string, config LimaConfig) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("error marshaling config: %w", err)
	}

	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	return nil
}

func mapServer(server Instance) (*types.Server, error) {
	ip, err := getIPFromVM(server.Name)
	if err != nil {
		return nil, fmt.Errorf("error getting IP address for %s: %w", server.Name, err)
	}
	return &types.Server{
		ObjectMeta: types.ObjectMeta{
			Name: server.Name,
		},
		Spec: types.ServerSpec{
			Image: server.Config.Images[0].Location,
		},
		Status: types.ServerStatus{
			PublicNet: &types.PublicNet{
				FQDN: server.Name,
				IPv4: &struct {
					IP string `json:"ip"`
				}{
					IP: ip,
				},
			},
		},
	}, nil
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

// getIPFromVM returns the IP address of the VM with the given name
func getIPFromVM(vmName string) (string, error) {
	listCmd := exec.Command("limactl", "shell", vmName, "ip", "-json", "addr", "show", "dev", "lima0")
	output, err := listCmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error listing VMs: %v, output: %s", err, output)
	}
	type AddressInfo struct {
		Family string `json:"family"`
		Local  string `json:"local"`
	}
	type NetworkInterface struct {
		AddrInfo []AddressInfo `json:"addr_info"`
	}

	var interfaces []NetworkInterface
	err = json.Unmarshal(output, &interfaces)
	if err != nil {
		return "", fmt.Errorf("error unmarshalling JSON: %w, output: %s", err, output)
	}

	return interfaces[0].AddrInfo[0].Local, nil
}
