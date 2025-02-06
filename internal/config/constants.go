package config

// Owner related constants
const (
	// DefaultOwner is the default owner
	DefaultOwner = "NO OWNER SET"

	// DefaultOrganization is the default organization
	DefaultOrganization = "NO ORGANIZATION SET"

	// DefaultEmail is the default email
	DefaultEmail = "NO EMAIL SET"
)

const (
	// ToolName is the name of the tool, used for directory names and logging
	ToolName = "storctl"

	// DefaultConfigDir is the default directory for all tool-related files
	DefaultConfigDir = ".storctl"

	// DefaultTemplateDir is the default directory for lab templates
	DefaultTemplateDir = "templates"

	// DefaultKeysDir is the default directory for storing SSH keys
	DefaultKeysDir = "keys"

	// ConfigFileName is the name of the configuration file
	ConfigFileName = "config.yaml"

	// DefaultAdminUser is the default admin user
	DefaultAdminUser = "ansible"

	// DefaultLabBucket is the default bucket for storing labs
	DefaultLabBucket = "labs"

	// DefaultLabStorageFile is the default file for storing labs
	DefaultLabStorageFile = "labs.db"

	// DefaultAnsibleDir is the default directory for storing ansible files
	DefaultAnsibleDir = "ansible"

	// DefaultAnsibleConfigFile is the default ansible config file
	DefaultAnsibleConfigFile = "ansible.cfg"

	// DefaultAnsibleInventoryFile is the default ansible inventory file
	DefaultAnsibleInventoryFile = "inventory"

	// DefaultAnsiblePlaybookFile is the default ansible playbook file
	DefaultAnsiblePlaybookFile = "site.yml"

	// DefaultAnsibleExtraVarsFile is the default ansible extra vars file
	DefaultAnsibleExtraVarsFile = "extra_vars.yml"

	// DefaultLimaDir is the default directory for storing lima VM configs
	DefaultLimaDir = "lima"
)

// Provider related constants
const (
	// DefaultLocalProvider is the default provider for a local machine
	DefaultLocalProvider = "lima"

	// DefaultLocalLocation is the default location
	DefaultLocalLocation = "local"

	// DefaultCloudProvider is the default cloud provider
	DefaultCloudProvider = "hetzner"

	// DefaultCloudLocation is the default cloud location
	DefaultCloudLocation = "nbg1"

	// DefaultDomain is the default domain
	DefaultDomain = "aistorlabs.com"

	// DefaultAdminKeyName is the default SSH key name
	DefaultAdminKeyName = "aistor-admin"

	// DefaultImage is the default image
	DefaultImage = "ubuntu-24.04"

	// DefaultServerType is the default server type
	DefaultServerType = "cx22"

	// DefaultToken is the default token
	DefaultToken = "NO TOKEN SET"

	// DefaultCredentials is the default credentials
	DefaultCredentials = "NOT USED WITH THIS PROVIDER"
)

// DNS related constants
const (
	// DefaultDNSProvider is the default DNS provider
	DefaultDNSProvider = "cloudflare"

	// DefaultDNSZoneID is the default DNS zone ID
	DefaultDNSZoneID = "NO ZONE ID SET"

	// DefaultDNSToken is the default DNS token
	DefaultDNSToken = "NO TOKEN SET"
)

// Time related constants
const (
	// DefaultTimeout is the default timeout for operations
	DefaultTimeout = "30s"

	// DefaultKeyTTL is the default time-to-live for SSH keys
	DefaultTTL = "1h"
)

// Volume related constants
const (
	// DefaultVolumeSize is the default size for volumes
	DefaultVolumeSize = 100

	// DefaultVolumeFormat is the default format for volumes
	DefaultVolumeFormat = "xfs"

	// DefaultVolumeAutomount is the default automount for volumes
	DefaultVolumeAutomount = false
)

// cloud-init
const (
	// DefaultCloudInitUserData is the default user data for cloud-init
	DefaultCloudInitUserData = `#cloud-config
users:
- name: ansible
  gecos: Ansible User
  groups: users,admin,wheel,sudo
  sudo: ALL=(ALL) NOPASSWD:ALL
  shell: /bin/bash
  ssh_authorized_keys:
  - %s

package_update: true
package_upgrade: true

power_state:
  mode: reboot
  message: Rebooting after package upgrades
  condition: test -f /var/run/reboot-required
`
)
