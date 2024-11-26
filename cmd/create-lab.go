package cmd

import (
	"github.com/pavelanni/labshop/internal/types"
	"github.com/pavelanni/labshop/internal/util/labelutil"
	"github.com/pavelanni/labshop/internal/util/timeutil"
	"github.com/spf13/cobra"
)

func NewCreateLabCmd() *cobra.Command {
	var (
		template string
	)

	cmd := &cobra.Command{
		Use:   "lab [name]",
		Short: "Create a new lab environment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			labName := args[0]
			return providerSvc.CreateLab(labName, template)
		},
	}

	cmd.Flags().StringVar(&template, "template", "", "Lab template to use (required)")
	if err := cmd.MarkFlagRequired("template"); err != nil {
		panic(err)
	}

	return cmd
}

func createLab(lab *types.Resource) error {
	lab.Metadata["owner"] = labelutil.SanitizeValue(cfg.Owner)
	lab.Metadata["organization"] = labelutil.SanitizeValue(cfg.Organization)
	lab.Metadata["email"] = labelutil.SanitizeValue(cfg.Email)

	servers := lab.Spec["servers"].([]map[string]any)
	for _, server := range servers {
		r := &types.Resource{
			Metadata: map[string]any{
				"name": server["name"],
				"labels": map[string]string{
					"delete_after": timeutil.FormatDeleteAfter(timeutil.TtlToDeleteAfter(server["ttl"].(string))),
					"owner":        labelutil.SanitizeValue(cfg.Owner),
				},
			},
			Spec: map[string]any{
				"provider":   server["provider"],
				"location":   server["location"],
				"serverType": server["serverType"],
				"ttl":        server["ttl"],
				"image":      server["image"],
				"keys":       server["keys"],
			},
		}
		if err := createServer(r); err != nil {
			return err
		}
	}
	return nil
}
