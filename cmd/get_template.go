package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pavelanni/storctl/internal/config"
	"github.com/pavelanni/storctl/internal/types"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/yaml"
)

type ServerConfig struct {
	CPUs   int
	Memory string
	Disk   string
}

var serverTypes = map[string]ServerConfig{
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

func NewGetTemplateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "template [template-id]",
		Short: "Get information about templates",
		Long:  `Display a list of all active templates or detailed information about a specific template`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return listTemplates()
			}
			return getTemplate(args[0])
		},
	}

	return cmd
}

func listTemplates() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	templateDir := filepath.Join(homeDir, ".storctl", config.DefaultTemplateDir)
	if _, err := os.Stat(templateDir); os.IsNotExist(err) {
		return fmt.Errorf("template directory %s does not exist", templateDir)
	}
	files, err := os.ReadDir(templateDir)
	if err != nil {
		return err
	}
	fmt.Println("Available templates:")
	for _, file := range files {
		if !file.IsDir() {
			if strings.HasSuffix(file.Name(), ".yaml") {
				fmt.Println(file.Name())
			}
		}
	}
	return nil
}

func getTemplate(templateID string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	templateDir := filepath.Join(homeDir, ".storctl", config.DefaultTemplateDir)
	if _, err := os.Stat(templateDir); os.IsNotExist(err) {
		return fmt.Errorf("template directory %s does not exist", templateDir)
	}
	if !strings.HasSuffix(templateID, ".yaml") {
		templateID = templateID + ".yaml"
	}
	templateFile := filepath.Join(templateDir, templateID)
	if _, err := os.Stat(templateFile); os.IsNotExist(err) {
		return fmt.Errorf("template file %s does not exist", templateFile)
	}
	template, err := os.ReadFile(templateFile)
	if err != nil {
		return err
	}

	decoder := yaml.NewYAMLOrJSONDecoder(bytes.NewBuffer(template), 4096)
	resource := &types.Resource{}
	if err := decoder.Decode(resource); err != nil {
		return fmt.Errorf("error decoding YAML: %w", err)
	}
	if resource.Kind != "Lab" {
		return fmt.Errorf("template is not a Lab")
	}
	lab := &types.Lab{}
	if err := convertToStruct(resource, lab); err != nil {
		return fmt.Errorf("error parsing Lab spec: %w", err)
	}

	fmt.Printf("Lab template: %s\n", strings.TrimSuffix(templateID, ".yaml"))
	for _, server := range lab.Spec.Servers {
		serverConfig, ok := serverTypes[server.ServerType]
		if !ok {
			serverConfig = ServerConfig{
				CPUs:   0,
				Memory: "unknown",
				Disk:   "unknown",
			}
		}
		fmt.Printf("  Server: %s, Type: %s, Cores: %d, Memory: %s, Disk: %s\n",
			server.Name,
			server.ServerType,
			serverConfig.CPUs,
			serverConfig.Memory,
			serverConfig.Disk)
	}
	for _, volume := range lab.Spec.Volumes {
		fmt.Printf("  Volume: %s, Size: %dGB\n",
			volume.Name,
			volume.Size)
	}
	return nil

}
