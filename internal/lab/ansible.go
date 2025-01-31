package lab

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
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
	ansibleUser := config.DefaultAdminUser
	ansibleSSHPrivateKeyFile := filepath.Join(homeDir, config.DefaultConfigDir, config.DefaultKeysDir, strings.Join([]string{lab.ObjectMeta.Name, "admin"}, "-"))
	if err != nil {
		return err
	}
	if m.Provider.Name() == "lima" {
		lab.Spec.CertManager = false
		lab.Spec.LetsEncrypt = "none"
		ansibleUser = os.Getenv("USER")
		ansibleSSHPrivateKeyFile = filepath.Join(homeDir, ".lima", "_config", "user")
	}

	allVars := map[string]any{
		"ansible_user":                 ansibleUser,
		"ansible_ssh_private_key_file": ansibleSSHPrivateKeyFile,
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
		m.Logger.Info("Generating Ansible inventory",
			"lab", lab.ObjectMeta.Name,
			"server_count", len(lab.Status.Servers))
		m.Logger.Debug("Adding server to inventory",
			"hostname", server.Status.PublicNet.FQDN,
			"cloud name", server.ObjectMeta.Name)
		if strings.HasSuffix(server.ObjectMeta.Name, "cp") {
			controlPlaneGroup.Hosts[server.Status.PublicNet.FQDN] = Host{
				AnsibleHost: server.Status.PublicNet.IPv4.IP,
			}
		} else {
			workerGroup.Hosts[server.Status.PublicNet.FQDN] = Host{
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
	m.Logger.Info("Creating Ansible inventory file", "file", ansibleInventoryFile)
	lab.Spec.Ansible.Inventory = ansibleInventoryFile
	err = m.Storage.Save(lab)
	if err != nil {
		return err
	}
	return os.WriteFile(ansibleInventoryFile, jsonData, 0644)
}

func (m *ManagerSvc) RunAnsiblePlaybook(lab *types.Lab) error {
	if lab.Spec.Ansible.Playbook == "" {
		return fmt.Errorf("ansible playbook not set")
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	ansiblePlaybookFile := filepath.Join(homeDir,
		config.DefaultConfigDir,
		config.DefaultAnsibleDir,
		"playbooks",
		lab.Spec.Ansible.Playbook)
	ansibleInventoryFile := lab.Spec.Ansible.Inventory
	m.Logger.Info("Running Ansible playbook", "playbook", ansiblePlaybookFile, "inventory", ansibleInventoryFile)
	if err := checkAnsibleAvailable(); err != nil {
		return err
	}
	args := []string{
		"-i", ansibleInventoryFile,
		ansiblePlaybookFile,
	}

	lab.Spec.Ansible.Playbook = ansiblePlaybookFile
	err = m.Storage.Save(lab)
	if err != nil {
		return err
	}
	cmd := exec.Command("ansible-playbook", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// checkAnsibleAvailable verifies that ansible-playbook is installed
func checkAnsibleAvailable() error {
	_, err := exec.LookPath("ansible-playbook")
	if err != nil {
		return fmt.Errorf("ansible-playbook not found in PATH: %w", err)
	}
	return nil
}
