package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/pavelanni/storctl/internal/types"
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
	resourceName := resource.ObjectMeta.Name
	if resourceName == "" {
		return fmt.Errorf("resource name is required")
	}
	switch resource.Kind {
	case "Server":
		if status := providerSvc.DeleteServer(resourceName, skipTimeCheck); status.Error != nil {
			return fmt.Errorf("failed to delete server: %w", status.Error)
		}
		return nil
	case "Volume":
		if status := providerSvc.DeleteVolume(resourceName, skipTimeCheck); status.Error != nil {
			return fmt.Errorf("failed to delete volume: %w", status.Error)
		}
		return nil
	case "Key":
		if status := providerSvc.DeleteSSHKey(resourceName, skipTimeCheck); status.Error != nil {
			return fmt.Errorf("failed to delete key: %w", status.Error)
		}
		return nil
	case "Lab":
		if status := providerSvc.DeleteLab(resourceName, skipTimeCheck); status.Error != nil {
			return fmt.Errorf("failed to delete lab: %w", status.Error)
		}
		return nil
	default:
		return fmt.Errorf("unknown resource kind: %s", resource.Kind)
	}
}

func askForConfirmation(resource *types.Resource) bool {
	resourceName := resource.ObjectMeta.Name
	if resourceName == "" {
		return false
	}
	resourceKind := resource.Kind
	fmt.Printf("Are you sure you want to delete %s %s? [y/N] ", resourceKind, resourceName)
	var response string
	fmt.Scanf("%s", &response)
	return response == "y" || response == "Y"
}

func askForConfirmationSimple(kind, name string) bool {
	fmt.Printf("Are you sure you want to delete %s %s? [y/N] ", kind, name)
	var response string
	fmt.Scanf("%s", &response)
	return response == "y" || response == "Y"
}
