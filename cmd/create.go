package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/pavelanni/labshop/internal/types"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func NewCreateCmd() *cobra.Command {
	var filename string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create resources (key, server, volume, lab)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if filename != "" {
				return createFromFile(filename)
			}
			return fmt.Errorf("either -f flag or a resource type must be specified")
		},
	}

	// Add -f flag
	cmd.Flags().StringVarP(&filename, "filename", "f", "", "Path to the YAML manifest file")

	// Add subcommands for direct resource creation
	cmd.AddCommand(
		NewCreateKeyCmd(),
		NewCreateServerCmd(),
		NewCreateVolumeCmd(),
		NewCreateLabCmd(),
	)

	return cmd
}

func createFromFile(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	// Handle multiple documents in YAML
	decoder := yaml.NewYAMLOrJSONDecoder(bytes.NewBuffer(data), 4096)
	for {
		resource := &types.Resource{}
		if err := decoder.Decode(resource); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("error decoding YAML: %w", err)
		}

		if err := processResource(resource); err != nil {
			return err
		}
	}

	return nil
}

func processResource(resource *types.Resource) error {
	switch resource.Kind {
	case "Server":
		return createServer(resource)
	case "Volume":
		return createVolume(resource)
	case "Key":
		return createKey(resource)
	case "Lab":
		return createLab(resource)
	default:
		return fmt.Errorf("unknown resource kind: %s", resource.Kind)
	}
}
