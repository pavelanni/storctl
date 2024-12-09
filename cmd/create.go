package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/pavelanni/storctl/internal/types"
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
		server := types.Server{
			TypeMeta:   resource.TypeMeta,
			ObjectMeta: resource.ObjectMeta,
		}
		// Convert map[string]interface{} to ServerSpec using json marshaling
		if err := convertToStruct(resource.Spec, &server.Spec); err != nil {
			return fmt.Errorf("error parsing Server spec: %w", err)
		}
		_, err := createServer(&server)
		return err
	case "Volume":
		volume := types.Volume{
			TypeMeta:   resource.TypeMeta,
			ObjectMeta: resource.ObjectMeta,
		}
		if err := convertToStruct(resource.Spec, &volume.Spec); err != nil {
			return fmt.Errorf("error parsing Volume spec: %w", err)
		}
		return createVolume(&volume)
	case "Key":
		key := types.SSHKey{
			TypeMeta:   resource.TypeMeta,
			ObjectMeta: resource.ObjectMeta,
		}
		if err := convertToStruct(resource.Spec, &key.Spec); err != nil {
			return fmt.Errorf("error parsing Key spec: %w", err)
		}
		_, err := createKey(&key)
		return err
	case "Lab":
		lab := &types.Lab{
			TypeMeta:   resource.TypeMeta,
			ObjectMeta: resource.ObjectMeta,
		}
		if err := convertToStruct(resource.Spec, &lab.Spec); err != nil {
			return fmt.Errorf("error parsing Lab spec: %w", err)
		}
		_, err := createLab(lab)
		if err != nil {
			return err
		}
		return nil
	default:
		return fmt.Errorf("unknown resource kind: %s", resource.Kind)
	}
}

// Helper function to convert map[string]interface{} to a struct
func convertToStruct(in interface{}, out interface{}) error {
	jsonBytes, err := json.Marshal(in)
	if err != nil {
		return err
	}
	return json.Unmarshal(jsonBytes, out)
}
