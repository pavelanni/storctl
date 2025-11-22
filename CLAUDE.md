# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project overview

storctl is a CLI tool for managing demo and lab environments for MinIO AIStor testing, training, and demonstrations. It supports both local deployment (using Lima VMs on macOS) and cloud deployment (currently Hetzner Cloud). The tool manages the complete lifecycle of lab environments including servers, volumes, DNS records, SSH keys, and automated Kubernetes + AIStor installation via Ansible.

## Architecture

### Core concepts

The codebase follows a provider pattern where different infrastructure providers (Lima for local VMs, Hetzner for cloud) implement a common `CloudProvider` interface. Resources follow a Kubernetes-style API model with TypeMeta, ObjectMeta, Spec, and Status fields.

### Key packages

- `cmd/` - Cobra CLI commands organized by action (create, get, delete, install)
- `internal/types` - Core resource types (Lab, Server, Volume, SSHKey) with Kubernetes-style structure
- `internal/provider` - Provider interface and implementations
  - `provider/hetzner` - Hetzner Cloud provider implementation
  - `provider/lima` - Lima VM provider for local development
  - `provider/factory.go` - Factory function to create providers based on config
- `internal/lab` - Lab manager that orchestrates multi-resource operations
  - Lab storage using BoltDB (bbolt)
  - Ansible inventory generation and playbook execution
  - Provider-specific lab creation/deletion logic
- `internal/config` - Configuration management and constants
- `internal/ssh` - SSH key management
- `internal/dns` - DNS management (Cloudflare)
- `assets/` - Embedded Ansible playbooks and templates using Go embed

### Resource model

All resources follow Kubernetes-style YAML manifests with:

- `apiVersion: v1`
- `kind: Lab|Server|Volume|SSHKey`
- `metadata:` (name, labels)
- `spec:` (resource specification)
- `status:` (runtime status)

### Lab lifecycle

1. Lab creation (`Create`) creates servers, volumes, and SSH keys via provider
1. For Hetzner: creates lab-specific SSH key, waits for servers to be SSH-ready
1. For Lima: creates volumes first, then servers with attached volumes
1. Lab data stored in local BoltDB at `~/.storctl/labs.db`
1. `SyncLabs()` fetches labs from provider by querying servers with `lab_name` label
1. `install lab` generates Ansible inventory and runs embedded playbooks to install K3s + AIStor

## Development commands

### Build

```bash
go build -o storctl .
```

### Run tests

```bash
# Run all tests
go test ./...

# Run tests for specific package
go test ./internal/lab
go test ./cmd

# Run tests with verbose output
go test -v ./...

# Run specific test
go test -run TestCreateLab ./cmd
```

### Release

Uses goreleaser for building multi-platform binaries:

```bash
goreleaser release --snapshot --clean
```

Version info is injected at build time via ldflags (see `.goreleaser.yaml`):

- `internal/version.Version`
- `internal/version.Commit`
- `internal/version.Date`

## Configuration

Configuration stored at `~/.storctl/`:

- `config.yaml` - Main config (providers, DNS, credentials)
- `labs.db` - BoltDB database for lab metadata
- `keys/` - SSH keys
- `ansible/` - Generated Ansible inventory files
- `lima/` - Lima VM configs
- `templates/` - Lab templates

## Provider interface

When adding a new provider, implement the `CloudProvider` interface in `internal/provider/types.go`:

- Server operations: CreateServer, GetServer, ListServers, AllServers, DeleteServer
- Volume operations: CreateVolume, GetVolume, ListVolumes, AllVolumes, DeleteVolume
- SSH Key operations: CreateSSHKey, GetSSHKey, ListSSHKeys, AllSSHKeys, DeleteSSHKey

Update `internal/provider/factory.go` to add the new provider case.

## Important implementation details

### Time-to-live (TTL)

Resources support TTL for automatic cleanup. TTL is parsed in `internal/util/timeutil` and stored as `DeleteAfter` timestamp in Status. Default TTL is 1 hour (see `config.DefaultTTL`).

### Label selectors

Resources use labels for filtering and grouping. Lab name is stored as `lab_name` label on all child resources. Use `LabelSelector` in ListOpts for querying (format: `"key=value"`).

### Server readiness checking

Hetzner labs use `internal/util/serverchecker` to verify SSH connectivity before proceeding with Ansible. This prevents race conditions where cloud-init hasn't finished.

### Ansible integration

- Playbooks embedded via `assets/PlaybookFiles` (Go embed.FS)
- Inventory generated dynamically from lab servers in JSON format
- Different SSH users for Lima (current user) vs Hetzner (ansible user)
- Playbooks install K3s, DirectPV, Helm, and AIStor

### Storage

Labs stored in BoltDB bucket (default: "labs"). Storage operations in `internal/lab/lab.go`:

- `Storage.Save()` - Marshal lab to JSON and store by name
- `Storage.Get()` - Retrieve lab by name
- `Storage.Delete()` - Remove lab from bucket
- `SyncLabs()` - Refresh local storage from provider

## Testing approach

The codebase uses table-driven tests with mocks:

- Provider mocks in `internal/provider/mock/`
- Lab manager mocks in `internal/lab/mock/`
- SSH mocks in `internal/util/serverchecker/mock_ssh.go`

Example test structure in `cmd/create_lab_test.go` and `cmd/delete_lab_test.go`.
