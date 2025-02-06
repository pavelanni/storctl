package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pavelanni/storctl/internal/config"
	"github.com/pavelanni/storctl/internal/types"
	"github.com/pavelanni/storctl/internal/util/labelutil"
	"github.com/pavelanni/storctl/internal/util/timeutil"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func NewCreateLabCmd() *cobra.Command {
	var (
		template string
		name     string
		provider string
		location string
		ttl      string
		playbook string
	)

	cmd := &cobra.Command{
		Use:   "lab [name]",
		Short: "Create a new lab environment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name = args[0]
			lab, err := labFromTemplate(template, name, provider, location, ttl, playbook)
			if err != nil {
				return fmt.Errorf("error parsing lab template: %w", err)
			}
			if provider != "" {
				lab.Spec.Provider = provider // override the provider in the template
			}
			_, err = createLab(lab)
			if err != nil {
				return fmt.Errorf("error creating lab: %w", err)
			}
			return nil
		},
	}

	defaultTemplate := filepath.Join(os.Getenv("HOME"), config.DefaultConfigDir, config.DefaultTemplateDir, "lab.yaml")
	cmd.Flags().StringVar(&template, "template", defaultTemplate, "lab template to use")
	cmd.Flags().StringVar(&provider, "provider", config.DefaultLocalProvider, "provider to use")
	cmd.Flags().StringVar(&location, "location", config.DefaultLocalLocation, "location to use")
	cmd.Flags().StringVar(&ttl, "ttl", config.DefaultTTL, "ttl to use")
	cmd.Flags().StringVar(&playbook, "playbook", "site.yml", "playbook to use")

	return cmd
}

func createLab(lab *types.Lab) (*types.Lab, error) {
	lab.ObjectMeta.Labels["owner"] = labelutil.SanitizeValue(cfg.Owner)
	lab.ObjectMeta.Labels["organization"] = labelutil.SanitizeValue(cfg.Organization)
	lab.ObjectMeta.Labels["email"] = labelutil.SanitizeValue(cfg.Email)
	lab.ObjectMeta.Labels["lab_name"] = lab.ObjectMeta.Name
	ttl := lab.Spec.TTL
	if ttl == "" {
		ttl = config.DefaultTTL
	}
	duration, err := timeutil.TtlToDuration(ttl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ttl: %w", err)
	}
	lab.ObjectMeta.Labels["delete_after"] = timeutil.FormatDeleteAfter(time.Now().Add(duration))

	err = initProvider(lab.Spec.Provider)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize provider: %w", err)
	}
	err = initLabManager()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize lab manager: %w", err)
	}

	fmt.Printf("Lab %s: Creating lab resources using provider %s...\n", lab.ObjectMeta.Name, lab.Spec.Provider)
	labSvc.Logger.Info("Creating new lab",
		"name", lab.ObjectMeta.Name,
		"nodes", len(lab.Spec.Servers))
	labSvc.Logger.Debug("Lab configuration", "lab", lab) // Detailed config for debugging
	if err := labSvc.Create(lab); err != nil {           // labSvc is a package variable created in root.go
		return nil, err
	}
	// get the lab again to get the status
	labUpdated, err := labSvc.Get(lab.ObjectMeta.Name)
	if err != nil {
		return nil, err
	}
	lab.Status = labUpdated.Status

	if lab.Spec.Provider != "lima" { // we don't need DNS records for local VMs
		fmt.Printf("Lab %s: Creating DNS records...\n", lab.ObjectMeta.Name)
		if err := addDNSRecords(lab); err != nil {
			return nil, err
		}
	}
	fmt.Printf("Lab %s: Creating ansible inventory file...\n", lab.ObjectMeta.Name)
	err = labSvc.CreateAnsibleInventoryFile(lab)
	if err != nil {
		return nil, err
	}
	if lab.Spec.Ansible.Playbook != "" {
		fmt.Printf("Lab %s: Running Ansible playbook %s...\n", lab.ObjectMeta.Name, lab.Spec.Ansible.Playbook)
	} else {
		fmt.Printf("Lab %s: No playbook specified. Skipping Ansible configuration.\n", lab.ObjectMeta.Name)
	}

	if lab.Spec.Ansible.Playbook != "" {
		err = labSvc.RunAnsiblePlaybook(lab)
		if err != nil {
			return nil, err
		}
	}
	return lab, nil
}

func labFromTemplate(template, name, provider, location, ttl, playbook string) (*types.Lab, error) {
	// Check if the template file exists
	if _, err := os.Stat(template); os.IsNotExist(err) {
		// Check if it exists in the default template directory
		tmpl := filepath.Join(os.Getenv("HOME"), config.DefaultConfigDir, config.DefaultTemplateDir, template)
		if _, err := os.Stat(tmpl); os.IsNotExist(err) {
			return nil, fmt.Errorf("template file does not exist: %s", tmpl)
		}
		template = tmpl
	}
	data, err := os.ReadFile(template)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	decoder := yaml.NewYAMLOrJSONDecoder(bytes.NewBuffer(data), 4096)
	lab := &types.Lab{}
	if err := decoder.Decode(lab); err != nil {
		return nil, fmt.Errorf("error decoding YAML: %w", err)
	}

	// Set values with defaults using the common pattern:
	lab.ObjectMeta.Name = name
	lab.Spec.Provider = defaultIfEmpty(lab.Spec.Provider, provider)
	lab.Spec.Location = defaultIfEmpty(lab.Spec.Location, location)
	lab.Spec.TTL = defaultIfEmpty(lab.Spec.TTL, ttl)
	lab.Spec.CertManager = defaultIfEmptyBool(lab.Spec.CertManager, true)
	lab.Spec.LetsEncrypt = defaultIfEmpty(lab.Spec.LetsEncrypt, "staging")
	lab.Spec.Ansible.Playbook = defaultIfEmpty(lab.Spec.Ansible.Playbook, playbook)
	lab.Spec.Ansible.User = defaultIfEmpty(lab.Spec.Ansible.User, config.DefaultAdminUser)
	lab.Spec.Ansible.Inventory = defaultIfEmpty(lab.Spec.Ansible.Inventory, "inventory.json")
	lab.Spec.Ansible.ConfigFile = defaultIfEmpty(lab.Spec.Ansible.ConfigFile, "ansible.cfg")

	return lab, nil
}

func addDNSRecords(lab *types.Lab) error {
	labName, ok := lab.ObjectMeta.Labels["lab_name"]
	if !ok {
		labName = "no-lab"
	}
	labName = strings.ToLower(labName)
	for i, server := range lab.Status.Servers {
		serverName := strings.ToLower(server.Name)
		// remove the leading labName with "-" from the serverName
		serverName = strings.TrimPrefix(serverName, labName+"-")
		err := dnsSvc.AddRecord(cfg.DNS.ZoneID,
			strings.Join([]string{serverName, labName}, "."),
			"A",
			server.Status.PublicNet.IPv4.IP,
			false)
		if err != nil {
			return err
		}
		lab.Status.Servers[i].Status.PublicNet.FQDN = strings.Join([]string{serverName, labName, cfg.DNS.Domain}, ".")
	}
	// Add a DNS record for 'aistor.' using the IP of the control plane server
	cpPublicNet := lab.Status.Servers[0].Status.PublicNet
	if err := dnsSvc.AddRecord(cfg.DNS.ZoneID,
		strings.Join([]string{"aistor", labName}, "."),
		"A",
		cpPublicNet.IPv4.IP,
		false); err != nil {
		return err
	}
	return nil
}

// Helper function to handle default string values
func defaultIfEmpty(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}

// Same for bool values (we'll use generics later, if necessary)
func defaultIfEmptyBool(value, defaultValue bool) bool {
	if !value {
		return defaultValue
	}
	return value
}
