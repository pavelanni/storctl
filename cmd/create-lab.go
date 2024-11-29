package cmd

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pavelanni/labshop/internal/config"
	"github.com/pavelanni/labshop/internal/types"
	"github.com/pavelanni/labshop/internal/util/labelutil"
	"github.com/pavelanni/labshop/internal/util/serverchecker"
	"github.com/pavelanni/labshop/internal/util/timeutil"
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
	cmd.Flags().StringVar(&provider, "provider", cfg.Provider.Name, "provider to use")
	cmd.Flags().StringVar(&location, "location", cfg.Provider.Location, "location to use")
	cmd.Flags().StringVar(&ttl, "ttl", config.DefaultTTL, "ttl to use")

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
	lab.ObjectMeta.Labels["delete_after"] = timeutil.FormatDeleteAfter(timeutil.TtlToDeleteAfter(ttl))

	keyNames := []string{strings.Join([]string{lab.ObjectMeta.Name, "admin"}, "-")}
	// Create servers
	specServers := lab.Spec.Servers
	servers := make([]*types.Server, 0)
	for _, serverSpec := range specServers {
		s := &types.Server{
			TypeMeta: types.TypeMeta{
				Kind:       "Server",
				APIVersion: "v1",
			},
			ObjectMeta: types.ObjectMeta{
				Name:   strings.Join([]string{lab.ObjectMeta.Name, serverSpec.Name}, "-"),
				Labels: lab.ObjectMeta.Labels,
			},
			Spec: types.ServerSpec{
				Location:    lab.Spec.Location,
				Type:        serverSpec.Type,
				TTL:         ttl,
				Image:       serverSpec.Image,
				SSHKeyNames: keyNames,
			},
		}
		result, err := createServer(s)
		if err != nil {
			return nil, err
		}
		if err := addDNSRecord(result); err != nil {
			return nil, err
		}
		servers = append(servers, result)
	}
	// Wait for servers to be ready
	results, err := serverchecker.CheckServers(context.Background(), servers)
	if err != nil {
		return nil, err
	}
	for _, result := range results {
		fmt.Printf("Server %s: %+v\n", result.Server.ObjectMeta.Name, result)
		if !result.Ready {
			return nil, fmt.Errorf("server %s not ready", result.Server.ObjectMeta.Name)
		}
	}

	// Create volumes
	volumes := lab.Spec.Volumes
	for _, volumeSpec := range volumes {
		if !volumeSpec.Automount { // if not specified, default to false
			volumeSpec.Automount = config.DefaultVolumeAutomount
		}
		if volumeSpec.Format == "" { // if not specified, default to xfs
			volumeSpec.Format = config.DefaultVolumeFormat
		}
		v := &types.Volume{
			TypeMeta: types.TypeMeta{
				Kind:       "Volume",
				APIVersion: "v1",
			},
			ObjectMeta: types.ObjectMeta{
				Name:   strings.Join([]string{lab.ObjectMeta.Name, volumeSpec.Name}, "-"),
				Labels: lab.ObjectMeta.Labels,
			},
			Spec: types.VolumeSpec{
				Size:       volumeSpec.Size,
				ServerName: strings.Join([]string{lab.ObjectMeta.Name, volumeSpec.Server}, "-"),
				Automount:  volumeSpec.Automount,
				Format:     volumeSpec.Format,
			},
		}
		if err := createVolume(v); err != nil {
			return nil, err
		}
	}
	return lab, nil
}

func labFromTemplate(template, name, provider, location, ttl string) (*types.Lab, error) {
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

func addDNSRecord(server *types.Server) error {
	labName, ok := server.ObjectMeta.Labels["lab_name"]
	if !ok {
		labName = "no-lab"
	}
	err := dnsSvc.AddRecord(cfg.DNS.ZoneID,
		strings.Join([]string{server.Name, labName}, "."),
		"A",
		server.Status.PublicNet.IPv4.IP,
		false)
	if err != nil {
		return err
	}
	return nil
}
