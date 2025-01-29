package lima

import (
	"fmt"
	"os/exec"
	"strings"
)

// createDisk creates a disk using limactl command
func createDisk(diskName, size string) error {
	// Check if disk already exists using limactl disk list
	listCmd := exec.Command("limactl", "disk", "list")
	output, err := listCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error listing disks: %v, output: %s", err, output)
	}

	// If disk is in the list, skip creation
	if strings.Contains(string(output), diskName) {
		fmt.Printf("Disk %s already exists, skipping creation\n", diskName)
		return nil
	}

	// Create disk using limactl disk create
	createCmd := exec.Command("limactl", "disk", "create", "--size", size, diskName)
	output, err = createCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error creating disk: %v, output: %s", err, output)
	}

	fmt.Printf("Created disk %s of size %s\n", diskName, size)
	return nil
}
