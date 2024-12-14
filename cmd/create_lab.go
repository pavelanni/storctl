package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pavelanni/storctl/internal/config"
	"github.com/pavelanni/storctl/internal/lab"
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
	)

	cmd := &cobra.Command{
		Use:   "lab [name]",
		Short: "Create a new lab environment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name = args[0]
			lab, err := labFromTemplate(template, name, provider, location, ttl)
			if err != nil {
				return fmt.Errorf("error parsing lab template: %w", err)
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
	cmd.Flags().StringVar(&provider, "provider", config.DefaultProvider, "provider to use")
	cmd.Flags().StringVar(&location, "location", config.DefaultLocation, "location to use")
	cmd.Flags().StringVar(&ttl, "ttl", config.DefaultTTL, "ttl to use")

	return cmd
}

func createLab(newLab *types.Lab) (*types.Lab, error) {
	newLab.ObjectMeta.Labels["owner"] = labelutil.SanitizeValue(cfg.Owner)
	newLab.ObjectMeta.Labels["organization"] = labelutil.SanitizeValue(cfg.Organization)
	newLab.ObjectMeta.Labels["email"] = labelutil.SanitizeValue(cfg.Email)
	newLab.ObjectMeta.Labels["lab_name"] = newLab.ObjectMeta.Name
	ttl := newLab.Spec.TTL
	if ttl == "" {
		ttl = config.DefaultTTL
	}
	duration, err := timeutil.TtlToDuration(ttl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ttl: %w", err)
	}
	newLab.ObjectMeta.Labels["delete_after"] = timeutil.FormatDeleteAfter(time.Now().Add(duration))

	labManager, err := lab.NewManager(providerSvc, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create lab manager: %w", err)
	}
	if err := labManager.Create(newLab); err != nil {
		return nil, err
	}
	if err := addDNSRecords(newLab); err != nil {
		return nil, err
	}
	return newLab, nil
}

func labFromTemplate(template, name, provider, location, ttl string) (*types.Lab, error) {
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
	lab.ObjectMeta.Name = name
	lab.Spec.Provider = provider
	lab.Spec.Location = location
	lab.Spec.TTL = ttl
	return lab, nil
}

func addDNSRecords(lab *types.Lab) error {
	labName, ok := lab.ObjectMeta.Labels["lab_name"]
	if !ok {
		labName = "no-lab"
	}
	labName = strings.ToLower(labName)
	for _, server := range lab.Status.Servers {
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
	}
	// Add a DNS record for 'aistor.' using the IP of the control plane server
	cpPublicNet := lab.Status.Servers[0].Status.PublicNet
	if err := dnsSvc.AddRecord(cfg.DNS.ZoneID,
		strings.Join([]string{labName, "aistor"}, "."),
		"A",
		cpPublicNet.IPv4.IP,
		false); err != nil {
		return err
	}
	return nil
}