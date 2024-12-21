package lab

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/pavelanni/storctl/internal/config"
	"github.com/pavelanni/storctl/internal/types"
)

// Host represents a single server
type Host struct {
	AnsibleHost string            `json:"ansible_host"`
	Vars        map[string]string `json:"vars,omitempty"`
}

// HostGroup represents a group of servers
type HostGroup struct {
	Hosts map[string]Host `json:"hosts"`
	Vars  map[string]any  `json:"vars,omitempty"`
}

// Inventory represents the complete Ansible inventory structure
type Inventory struct {
	All struct {
		Children map[string]HostGroup `json:"children"`
		Vars     map[string]any       `json:"vars,omitempty"`
	} `json:"all"`
}

func (m *ManagerSvc) CreateAnsibleInventoryFile(lab *types.Lab) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	allVars := map[string]any{
		"ansible_user":                 config.DefaultAdminUser,
		"ansible_ssh_private_key_file": filepath.Join(homeDir, config.DefaultConfigDir, config.DefaultKeysDir, strings.Join([]string{lab.ObjectMeta.Name, "admin"}, "-")),
		"ansible_ssh_common_args":      "-o StrictHostKeyChecking=no",
		"lab_name":                     lab.ObjectMeta.Name,
		"domain_name":                  config.DefaultDomain,
		"letsencrypt_environment":      lab.Spec.LetsEncrypt,
		"cert_manager_enable":          lab.Spec.CertManager,
	}
	ansibleInventoryFile := filepath.Join(homeDir, config.DefaultConfigDir, config.DefaultAnsibleDir, strings.Join([]string{lab.ObjectMeta.Name, "inventory.json"}, "-"))

	inventory := Inventory{
		All: struct {
			Children map[string]HostGroup `json:"children"`
			Vars     map[string]any       `json:"vars,omitempty"`
		}{
			Children: make(map[string]HostGroup),
			Vars:     allVars,
		},
	}
	controlPlaneGroup := HostGroup{
		Hosts: make(map[string]Host),
	}
	workerGroup := HostGroup{
		Hosts: make(map[string]Host),
	}
	for _, server := range lab.Status.Servers {
		if strings.HasSuffix(server.ObjectMeta.Name, "cp") {
			controlPlaneGroup.Hosts[server.ObjectMeta.Name] = Host{
				AnsibleHost: server.Status.PublicNet.IPv4.IP,
			}
		} else {
			workerGroup.Hosts[server.ObjectMeta.Name] = Host{
				AnsibleHost: server.Status.PublicNet.IPv4.IP,
			}
		}
	}
	inventory.All.Children["control_plane"] = controlPlaneGroup
	inventory.All.Children["nodes"] = workerGroup
	inventory.All.Vars = allVars

	jsonData, err := json.MarshalIndent(inventory, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(ansibleInventoryFile, jsonData, 0644)
}
