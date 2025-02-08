# storctl - AIStor Environment Manager

`storctl` is a command-line tool for managing demo and lab environments in cloud infrastructure or on the local host using virtual machines.
The main focus of this tool is on MinIO AIStor testing, training, and demonstration.

## Features

- Create and manage lab environments with multiple servers and volumes
- Manage DNS records with Cloudflare
- Use Lima virtual machines on macOS or Hetzner Cloud infrastructure (currently)
- Manage SSH keys to access cloud VMs
- Manage cloud resource lifecycle with TTL (Time To Live)
- Use YAML-based configuration and resource definitions similar to Kubernetes

## Installation

### Prerequisites

#### For both deployments

1. `kubectl` is installed. If it's not installed on your machine, follow these [instructions](https://kubernetes.io/docs/tasks/tools/#kubectl)

1. Krew is installed. If it's not installed, follow these [instructions](https://krew.sigs.k8s.io/docs/user-guide/setup/install/)

1. DirectPV plugin is installed. If it's not installed, follow these [instructions](https://min.io/docs/directpv/installation/#install-directpv-plugin-with-krew)

1. Helm is installed. If it's not installed on your machine, follow these [instructions](https://helm.sh/docs/intro/install/).
   On a Mac, the easiest way is to use `brew install helm`.

#### For local deployment

Local AIStor installation uses Lima to manage virtual machines, QEMU as a virtualization engine, and `socket_vmnet` for the network.
We have to use QEMU with `socket_vmnet` shared network to allow the VMs to talk to each other and being able to access the VMs from the host.

1. Install Lima.

   ```shell
   brew install lima
   ```

1. Install QEMU

   ```shell
   brew install qemu
   ```

1. Check if you have already installed Xcode command tools (which is very likely)

   ```shell
   xcode-select -p
   ```

   Expected output:

   ```none
   /Library/Developer/CommandLineTools
   ```

   If it's not installed, run:

   ```shell
   xcode-select --install
   ```

1. Build and install the network driver for `socket_vmnet`. The full instructions and explanation is provided on the official [Lima site](https://lima-vm.io/docs/config/network/#socket_vmnet).
   Here is a short version of it:

   ```shell
   # Install socket_vmnet as root from source to /opt/socket_vmnet
   # using instructions on https://github.com/lima-vm/socket_vmnet
   # This assumes that Xcode Command Line Tools are already installed
   git clone https://github.com/lima-vm/socket_vmnet
   cd socket_vmnet
   # Change "v1.2.1" to the actual latest release in https://github.com/lima-vm/socket_vmnet/releases
   git checkout v1.2.1
   make
   sudo make PREFIX=/opt/socket_vmnet install.bin

   # Set up the sudoers file for launching socket_vmnet from Lima
   limactl sudoers >etc_sudoers.d_lima
   less etc_sudoers.d_lima  # verify that the file looks correct
   sudo install -o root etc_sudoers.d_lima /etc/sudoers.d/lima
   rm etc_sudoers.d_lima
   ```

1. Note: Lima might give you an error message about the `docker.sock` file.
   In that case, just delete the file mentioned in the error message.

#### For cloud deployment

1. Get a Hetzner Cloud account and API token. Ask the Traning team for access to the MinIO shared project.

1. Get a Cloudflare account and API token (for DNS management) from the Training team.
   You don't need it if you prefer to use your own domain.

### Using released binaries (recommended)

Download binaries for your OS/arch from the [Releases](https://github.com/pavelanni/storctl/releases) page.

### Building from source

```bash
git clone https://github.com/pavelanni/storctl
cd storctl
go build -o storctl .
# Move the resulting binary to your PATH
mv storctl $HOME/.local/bin # or any other directory in your PATH
```

## Configuration

1. Initialize the configuration:

```bash
storctl init
```

This creates a default configuration directory at `~/.storctl` with the following structure:

- `config.yaml` -- Main configuration file
- `templates/` -- Lab environment templates
- `keys/` -- SSH key storage
- `ansible/` -- for Ansible playbooks and inventory files
- `lima/` -- for Lima configs

1. Edit the configuration file at `~/.storctl/config.yaml`:

```yaml
providers:
  - name: "hetzner"
    token: "your-hetzner-token"
    location: "nbg1" # EU locations: nbd1, fsn1, hel1; US locations: ash, hil; APAC locations: sin
  - name: "lima"

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
storctl create lab mylab --template lab.yaml

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

## Known issues

1. In multi-node configurations (with more than one worker node in the cluster) sometimes DirectPV doesn't discover
   drives on all nodes properly. Before starting using AIStor after installation, check DirectPV status with this command:

   ```shell
   kubectl directpv info
   ```

   If in the output you don't see all your nodes and drives, re-run the discovery and initialization commands:

   ```shell
   kubectl directpv discover
   kubectl directpv init drives.yaml --dangerous
   ```

   And check the status again with the `kubectl directpv info` command.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
