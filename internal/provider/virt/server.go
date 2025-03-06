package virt

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/digitalocean/go-libvirt"
	"github.com/pavelanni/storctl/internal/config"
	"github.com/pavelanni/storctl/internal/provider/options"
	"github.com/pavelanni/storctl/internal/types"
)

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

func (p *VirtProvider) CreateServer(opts options.ServerCreateOpts) (*types.Server, error) {
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
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("error getting home directory: %w", err)
	}
	virtDir := filepath.Join(homeDir, config.DefaultConfigDir, config.DefaultVirtDir)
	if _, err := os.Stat(virtDir); os.IsNotExist(err) {
		err = os.MkdirAll(virtDir, 0755)
		if err != nil {
			return nil, fmt.Errorf("error creating Virt directory: %w", err)
		}
	}
	virtConfigFile := filepath.Join(virtDir, opts.Name+".xml")
	server, ok := serverTypes[opts.Type]
	if !ok {
		return nil, fmt.Errorf("invalid server type: %s", opts.Type)
	}
	if len(opts.AdditionalDisks) > 0 {
		fmt.Printf("Additional disks in Virt provider from opts: %v\n", opts.AdditionalDisks)
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
	server.Arch = arch
	server.Image = opts.Image
	server.Name = opts.Name
	server.Role = opts.Role
	server.CPUs = opts.CPUs
	server.Memory = opts.Memory
	server.Disk = opts.Disk
}

func (p *VirtProvider) GetServer(name string) (*types.Server, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	cmd := exec.CommandContext(ctx, "virsh", "list", "--name")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("listing servers: %w", err)
	}

	// Empty output also means no server
	if len(strings.TrimSpace(string(output))) == 0 {
		return nil, nil
	}

}

func (p *VirtProvider) ListServers(opts options.ServerListOpts) ([]*types.Server, error) {

}

func (p *VirtProvider) AllServers() ([]*types.Server, error) {
	flags := libvirt.ConnectListDomainsActive | libvirt.ConnectListDomainsInactive
	domains, _, err := p.client.ConnectListAllDomains(1, flags)
	if err != nil {
		return nil, fmt.Errorf("listing servers: %w", err)
	}

	servers := []*types.Server{}
	for _, domain := range domains {
		servers = append(servers, &types.Server{
			ObjectMeta: types.ObjectMeta{
				Name: domain.Name(),
			},
		})
	}

}

func (p *VirtProvider) DeleteServer(name string, force bool) *types.ServerDeleteStatus {

}

func (p *VirtProvider) ServerToCreateOpts(server *types.Server) (options.ServerCreateOpts, error) {

}

func (p *VirtProvider) mapDomain(domain *libvirt.Domain) (*types.Server, error) {
	_, _, mem, nCpu, _, err := p.client.DomainGetInfo(*domain)
	if err != nil {
		return nil, fmt.Errorf("getting domain info: %w", err)
	}
	return &types.Server{
		ObjectMeta: types.ObjectMeta{
			Name: domain.Name,
		},
		Spec: types.ServerSpec{},
		Status: types.ServerStatus{
			Cores:  int(nCpu),
			Memory: float32(mem) / 1024 / 1024 / 1024,
		},
	}, nil
}

func getArchForImage(arch string) string {
	if arch == "amd64" || arch == "x86_64" {
		return "x86_64"
	}
	return arch
}

func getArchForArch(arch string) string {
	if arch == "amd64" || arch == "x86_64" {
		return "x86_64"
	}
	return arch
}

func getIPFromVM(vmName string) (string, error) {

}
