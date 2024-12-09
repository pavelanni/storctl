# storctl - AIStor Environment Manager

`storctl` is a command-line tool for managing demo and lab environments in cloud infrastructure. The main focus of this tool is on MinIO AIStor testing, training, and demonstration.

## Features

- Create and manage lab environments with multiple servers and volumes
- Manage DNS records with Cloudflare
- Use Hetzner Cloud infrastructure (currently)
- Manage SSH keys to access cloud VMs
- Manage resource lifecycle with TTL (Time To Live)
- Use YAML-based configuration and resource definitions similar to Kubernetes

## Installation

### Prerequisites

- Go 1.23 or later
- A Hetzner Cloud account and API token
- A Cloudflare account and API token (for DNS management)

### Building from source

```bash
git clone https://github.com/pavelanni/storctl
cd storctl
go build .
```

## Configuration

1. Initialize the configuration:

```bash
storctl init
```

This creates a default configuration directory at `~/.storctl` with the following structure:

- `config.yaml` - Main configuration file
- `templates/` - Lab environment templates
- `keys/` - SSH key storage

1. Edit the configuration file at `~/.storctl/config.yaml`:

```yaml
provider:
  token: "your-hetzner-token"
  name: "hetzner"
  location: "nbg1" # EU locations: nbd1, fsn1, hel1; US locations: ash, hil; APAC locations: sin

dns:
  provider: "cloudflare"
  token: "your-cloudflare-token"
  zone_id: "your-zone-id"
  domain: "aistorlabs.com" # feel free to use your own domain

email: "your-email@example.com"
organization: "your-organization"
owner: "your-name"
```

## Usage

### Basic Commands

```bash
# View current configuration
storctl config view

# Create a new lab environment
storctl create lab mylab

# List all labs
storctl get lab

# Get details about a specific lab
storctl get lab mylab

# Delete a lab
storctl delete lab mylab

# Create a new SSH key
storctl create key mykey

# Create a new server
storctl create server myserver

# Create a new volume
storctl create volume myvolume
```

### Using resource YAML files

You can also create resources using YAML definition files:

```bash
storctl create -f lab.yaml
storctl create -f server.yaml
storctl create -f volume.yaml
```

### Resource templates

Example lab template:

```yaml
apiVersion: v1
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
    serverType: cx22
    image: ubuntu-24.04
  - name: node-01
    serverType: cx22
    image: ubuntu-24.04
volumes:
  - name: volume-01
    server: node-01
    size: 100
    automount: false
    format: xfs
```

## Resource management

All resources support:

- Labels for organization and filtering
- TTL (Time To Live) for automatic cleanup
- Provider-specific configurations
- YAML/JSON manifest files

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
