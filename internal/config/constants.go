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
)

// Provider related constants
const (
	// DefaultProvider is the default provider
	DefaultProvider = "hetzner"

	// DefaultDomain is the default domain
	DefaultDomain = "aistorlabs.com"

	// DefaultLocation is the default location
	DefaultLocation = "nbg1"

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

// Lab template related constants
const (
	// DefaultLabTemplate is the default lab template
	DefaultLabTemplate = `apiVersion: v1
kind: Lab
metadata:
  name: aistor-lab
  labels:
    project: aistor
spec:
  ttl: 24h
  provider: hetzner
  location: nbg1
  servers:
  - name: cp
    type: cx22
    image: ubuntu-24.04
  - name: node-01
    type: cx22
    image: ubuntu-24.04
  volumes:
  - name: volume-01
    server: node-01
    size: 100
    automount: false
    format: xfs
  - name: volume-02
    server: node-01
    size: 100
    automount: false
    format: xfs
  - name: volume-03
    server: node-01
    size: 100
    automount: false
    format: xfs
  - name: volume-04
    server: node-01
    size: 100
    automount: false
    format: xfs
`
)
