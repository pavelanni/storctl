package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/pavelanni/storctl/internal/config"
	"github.com/pavelanni/storctl/internal/lab"
	"github.com/pavelanni/storctl/internal/provider"
)

type InstallLabOpts struct {
	LabName         string
	Inventory       string
	Playbook        string
	CreateInventory bool
}

func NewInstallLabCmd() *cobra.Command {
	opts := InstallLabOpts{}

	cmd := &cobra.Command{
		Use:   "lab LAB_NAME",
		Short: "Install software in a lab",
		Long:  "Install software in a lab by running Ansible playbook",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("lab name is required")
			}
			return installLab(args[0], opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Inventory, "inventory", "i", "", "path to the inventory file")
	cmd.Flags().StringVarP(&opts.Playbook, "playbook", "p", "site.yml", "path to the playbook file")
	//	cmd.Flags().BoolVarP(&opts.CreateInventory, "create-inventory", "c", false, "create the inventory file")

	return cmd
}

func installLab(labName string, opts InstallLabOpts) error {
	if opts.Inventory == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("error getting home directory: %w", err)
		}
		opts.Inventory = filepath.Join(homeDir,
			config.DefaultConfigDir,
			config.DefaultAnsibleDir,
			labName+"-inventory.json")
	}
	inventory := lab.Inventory{}
	data, err := os.ReadFile(opts.Inventory)
	if err != nil {
		return fmt.Errorf("error reading inventory file: %w", err)
	}
	if err := json.Unmarshal(data, &inventory); err != nil {
		return fmt.Errorf("error unmarshalling inventory file: %w", err)
	}
	if inventory.All.Vars["provider"] == nil {
		inventory.All.Vars["provider"] = config.DefaultLocalProvider
	}
	providerSvc, err := provider.NewProvider(*cfg, inventory.All.Vars["provider"].(string))
	if err != nil {
		return fmt.Errorf("error creating provider: %w", err)
	}
	labSvc, err := lab.NewManager(providerSvc, cfg)
	if err != nil {
		return fmt.Errorf("error creating lab manager: %w", err)
	}
	lab, err := labSvc.Get(labName) // get from the storage
	if err != nil {
		return fmt.Errorf("error getting lab: %w", err)
	}
	if opts.Playbook != "" {
		lab.Spec.Ansible.Playbook = opts.Playbook
	}
	if opts.Inventory != "" {
		lab.Spec.Ansible.Inventory = opts.Inventory
	}
	err = labSvc.RunAnsiblePlaybook(lab)
	if err != nil {
		return fmt.Errorf("error running ansible playbook: %w", err)
	}
	return nil
}
