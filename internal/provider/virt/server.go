package virt

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

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
	server, ok := serverTypes[opts.Type]
	if !ok {
		return nil, fmt.Errorf("invalid server type: %s", opts.Type)
	}
	arch := getArchForArch(p.arch)
	server.Arch = arch
	server.Image = opts.Image
	server.Name = opts.Name
	virtTemplateFile := filepath.Join(virtDir, config.DefaultTemplateDir, "domain-template.xml")
	xml := executeDomainTemplate(DomainConfig{
		Name:   opts.Name,
		Memory: server.Memory,
		Disk:   server.Disk,
		VCPU:   strconv.Itoa(server.CPUs),
	}, virtTemplateFile)

	virtConfigFile := filepath.Join(virtDir, opts.Name+".xml")
	err = os.WriteFile(virtConfigFile, []byte(xml), 0644)
	if err != nil {
		return nil, fmt.Errorf("error writing domain config file: %w", err)
	}
	domain, err := p.client.DomainCreateXML(xml, 0)
	if err != nil {
		return nil, fmt.Errorf("error creating domain: %w", err)
	}
	return p.mapDomain(&domain)
}

func (p *VirtProvider) GetServer(name string) (*types.Server, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	// Try to lookup the domain by name
	domain, err := p.client.DomainLookupByName(name)
	if err != nil {
		if strings.Contains(err.Error(), "Domain not found") {
			return nil, fmt.Errorf("server %s not found", name)
		}
		return nil, fmt.Errorf("looking up domain: %w", err)
	}

	// Map the domain to our Server type
	return p.mapDomain(&domain)
}

func (p *VirtProvider) ListServers(opts options.ServerListOpts) ([]*types.Server, error) {
	flags := libvirt.ConnectListDomainsActive | libvirt.ConnectListDomainsInactive
	var labName string

	if opts.ListOpts.LabelSelector != "" {
		label := opts.ListOpts.LabelSelector
		labName = strings.TrimPrefix(label, "lab_name=")
	}

	domains, _, err := p.client.ConnectListAllDomains(1, flags)
	if err != nil {
		return nil, fmt.Errorf("listing servers: %w", err)
	}

	servers := []*types.Server{}
	for _, domain := range domains {
		server, err := p.mapDomain(&domain)
		if err != nil {
			return nil, fmt.Errorf("mapping domain: %w", err)
		}
		if labName != "" && strings.HasPrefix(server.ObjectMeta.Name, labName) {
			servers = append(servers, server)
		}
	}

	return servers, nil
}

func (p *VirtProvider) AllServers() ([]*types.Server, error) {
	flags := libvirt.ConnectListDomainsActive | libvirt.ConnectListDomainsInactive
	domains, _, err := p.client.ConnectListAllDomains(1, flags)
	if err != nil {
		return nil, fmt.Errorf("listing servers: %w", err)
	}

	servers := []*types.Server{}
	for _, domain := range domains {
		server, err := p.mapDomain(&domain)
		if err != nil {
			return nil, fmt.Errorf("mapping domain: %w", err)
		}
		servers = append(servers, server)
	}

	return servers, nil
}

func (p *VirtProvider) DeleteServer(name string, force bool) *types.ServerDeleteStatus {
	domain, err := p.client.DomainLookupByName(name)
	if err != nil {
		return &types.ServerDeleteStatus{
			Deleted: false,
			Error:   err,
		}
	}
	err = p.client.DomainDestroy(domain)
	if err != nil {
		return &types.ServerDeleteStatus{
			Deleted: false,
			Error:   err,
		}
	}
	return &types.ServerDeleteStatus{
		Deleted: true,
	}

}

func (p *VirtProvider) ServerToCreateOpts(server *types.Server) (options.ServerCreateOpts, error) {
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
		Provider: "virt",
		SSHKeys:  sshKeys,
		UserData: cloudInitUserData,
	}, nil
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

func (p *VirtProvider) getIPFromDomain(domain *libvirt.Domain) (string, error) {
	// Use DomainInterfaceSourceLease to get IPs from DHCP leases
	interfaces, err := p.client.DomainInterfaceAddresses(*domain, uint32(libvirt.DomainInterfaceAddressesSrcLease), 0)
	if err != nil {
		return "", fmt.Errorf("listing interfaces: %w", err)
	}

	// Look through all interfaces
	for _, iface := range interfaces {
		// Look through all addresses on each interface
		for _, addr := range iface.Addrs {
			// Return the first IPv4 address we find
			if addr.Type == int32(libvirt.IPAddrTypeIpv4) {
				return addr.Addr, nil
			}
		}
	}
	return "", fmt.Errorf("no IP address found")
}

func executeDomainTemplate(config DomainConfig, vmTemplateFile string) string {
	tmpl, err := template.ParseFiles(vmTemplateFile)
	if err != nil {
		log.Fatalf("failed to parse template file: %v", err)
	}

	var result bytes.Buffer
	if err := tmpl.Execute(&result, config); err != nil {
		log.Fatalf("failed to execute template: %v", err)
	}

	return result.String()
}
