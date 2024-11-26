package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/pavelanni/labshop/internal/logger"
	"github.com/pavelanni/labshop/internal/types"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func NewDeleteCmd() *cobra.Command {
	var (
		filename      string
		assumeYes     bool
		skipTimeCheck bool
	)

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete resources (key, server, volume, lab)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if filename != "" {
				return deleteFromFile(filename, assumeYes, skipTimeCheck)
			}
			return fmt.Errorf("either -f flag or a resource type must be specified")
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&filename, "filename", "f", "", "Path to the YAML manifest file")
	cmd.Flags().BoolVarP(&assumeYes, "yes", "y", false, "Skip confirmation prompt")
	cmd.Flags().BoolVar(&skipTimeCheck, "force", false, "Skip resource creation time check")

	// Add subcommands for direct resource deletion
	cmd.AddCommand(
		NewDeleteSSHKeyCmd(),
		NewDeleteServerCmd(),
		NewDeleteVolumeCmd(),
		NewDeleteLabCmd(),
	)

	return cmd
}

func deleteFromFile(filename string, assumeYes, skipTimeCheck bool) error {
	logger.Info("processing delete operation from file",
		"filename", filename,
		"assumeYes", assumeYes,
		"skipTimeCheck", skipTimeCheck)
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

		if err := processDeleteResource(resource, assumeYes, skipTimeCheck); err != nil {
			return err
		}
	}

	return nil
}

func processDeleteResource(resource *types.Resource, assumeYes, skipTimeCheck bool) error {
	if !askForConfirmation(resource) {
		return nil
	}
	resourceName, ok := resource.Metadata["name"]
	if !ok {
		return fmt.Errorf("resource name is required")
	}
	switch resource.Kind {
	case "Server":
		if err := providerSvc.DeleteServer(resourceName.(string), skipTimeCheck); err != nil {
			return fmt.Errorf("failed to delete server: %w", err)
		}
		return nil
	case "Volume":
		if err := providerSvc.DeleteVolume(resourceName.(string), skipTimeCheck); err != nil {
			return fmt.Errorf("failed to delete volume: %w", err)
		}
		return nil
	case "Key":
		if err := providerSvc.DeleteSSHKey(resourceName.(string), skipTimeCheck); err != nil {
			return fmt.Errorf("failed to delete key: %w", err)
		}
		return nil
	case "Lab":
		if err := providerSvc.DeleteLab(resourceName.(string), skipTimeCheck); err != nil {
			return fmt.Errorf("failed to delete lab: %w", err)
		}
		return nil
	default:
		return fmt.Errorf("unknown resource kind: %s", resource.Kind)
	}
}

func askForConfirmation(resource *types.Resource) bool {
	resourceName, ok := resource.Metadata["name"]
	if !ok {
		return false
	}
	resourceKind := resource.Kind
	fmt.Printf("Are you sure you want to delete %s %s? [y/N] ", resourceKind, resourceName)
	var response string
	_, err := fmt.Scanln(&response)
	if err != nil {
		return false
	}
	return response == "y" || response == "Y"
}
