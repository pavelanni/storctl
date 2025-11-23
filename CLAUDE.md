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
- `internal/storage` - Storage abstraction layer
  - `storage/interface.go` - Storage interface (Save, Get, List, Delete, Close)
  - `storage/postgres/` - PostgreSQL implementation (production-ready)
  - `storage/local/` - BoltDB implementation (partial, needs Get/List/Delete)
- `internal/lab` - Lab manager that orchestrates multi-resource operations
  - Selects storage backend based on config (PostgreSQL or BoltDB)
  - Ansible inventory generation and playbook execution
  - Provider-specific lab creation/deletion logic
- `internal/config` - Configuration management and constants
- `internal/ssh` - SSH key management
- `internal/dns` - DNS management (Cloudflare)
- `assets/` - Embedded Ansible playbooks and templates using Go embed
- `migrations/` - Database migrations (PostgreSQL schema)
- `scripts/` - Helper scripts for PostgreSQL management

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
1. Lab data stored in configured backend (PostgreSQL or BoltDB)
   - PostgreSQL: Remote database with soft deletes, ACID transactions
   - BoltDB: Local file at `~/.storctl/labs.db` (embedded database)
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

- `config.yaml` - Main config (providers, DNS, credentials, storage backend)
- `labs.db` - BoltDB database for lab metadata (if using local storage)
- `postgres-data/` - PostgreSQL data directory (if using PostgreSQL storage with host mount)
- `backups/` - PostgreSQL database backups (created by `scripts/backup-postgres.sh`)
- `keys/` - SSH keys
- `ansible/` - Generated Ansible inventory files
- `lima/` - Lima VM configs
- `templates/` - Lab templates

## PostgreSQL setup (for development)

PostgreSQL can be run locally using Podman with persistent data storage:

**Start PostgreSQL:**
```bash
./scripts/start-postgres.sh
```

This script:
- Creates container if needed (or starts existing one)
- Uses host mount at `~/.storctl/postgres-data` for persistence
- Auto-applies migrations if database is empty
- Data survives container recreation and Podman machine recreation

**Other scripts:**
- `./scripts/stop-postgres.sh` - Stop PostgreSQL container
- `./scripts/backup-postgres.sh` - Backup database to `~/.storctl/backups/`
- `./scripts/restore-postgres.sh <file>` - Restore from backup

**Data persistence strategy:**
- PostgreSQL data stored on Mac filesystem (not in Podman volume)
- Path: `~/.storctl/postgres-data`
- Survives: container restarts, Podman machine recreation, macOS upgrades
- Does not survive: macOS reinstall (backup your home directory)

See `docs/POSTGRESQL_SETUP.md` for complete PostgreSQL documentation.

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

### Storage abstraction

The storage layer uses an interface-based architecture supporting multiple backends:

```go
// internal/storage/interface.go
type Storage interface {
    Save(lab *types.Lab) error
    Get(name string) (*types.Lab, error)
    List() ([]*types.Lab, error)
    Delete(name string) error
    Close() error
}
```

**Implementations:**
- **PostgreSQL** (`internal/storage/postgres/`) - Remote database, production-ready
  - Full CRUD operations
  - Soft deletes (sets `deleted_at` timestamp)
  - UPSERT on conflict (ON CONFLICT DO UPDATE)
  - Requires `_ "github.com/lib/pq"` driver import
  - Connection string: `host=%s port=%s dbname=%s user=%s password=%s sslmode=disable`
  - Schema in `migrations/001_initial.sql` (labs + audit_logs tables)

- **BoltDB** (`internal/storage/local/`) - Local embedded database
  - Save() implemented
  - Get/List/Delete need implementation (currently stubs)
  - No server required, single file at `~/.storctl/labs.db`

**Configuration** (`config.yaml`):
```yaml
storage:
  type: postgres  # or "local" for BoltDB
  postgres:
    host: localhost
    port: 5432
    database: storctl_dev
    user: postgres
    password: storctl
  local:
    path: ~/.storctl/labs.db
    bucket: labs
```

**Backend selection** in `internal/lab/lab.go`:
```go
func NewManager(provider provider.CloudProvider, cfg *config.Config) (*ManagerSvc, error) {
    var storage storage.Storage  // interface, NOT pointer to interface

    switch cfg.Storage.Type {
    case "postgres":
        storage, err = postgres.New(cfg)
    case "local", "":
        storage, err = local.New(cfg)
    }

    return &ManagerSvc{
        Storage: storage,  // No &, interfaces are already reference types
    }
}
```

**Important:** Never use pointer to interface (`*storage.Storage`). Interfaces are already reference types internally.

## Common patterns and best practices

### Interface + Service pattern

For packages that coordinate multiple dependencies:
```go
// Interface defines the contract
type Manager interface {
    Create(lab *types.Lab) error
    Get(labName string) (*types.Lab, error)
    // ...
}

// Service implementation holds dependencies
type ManagerSvc struct {
    Storage    storage.Storage       // interface, not *storage.Storage
    Provider   provider.CloudProvider
    Logger     *slog.Logger
}

// Compile-time check that ManagerSvc implements Manager
var _ Manager = (*ManagerSvc)(nil)
```

Use this pattern when you need to orchestrate multiple services. Don't use it for simple implementations that are just swapping behavior.

### Interfaces are reference types

**Never use pointers to interfaces:**
```go
// ❌ WRONG
type ManagerSvc struct {
    Storage *storage.Storage  // pointer to interface
}

// ✅ CORRECT
type ManagerSvc struct {
    Storage storage.Storage   // just the interface
}
```

Interfaces already contain pointers internally. Using `*interface` creates unnecessary indirection and is a code smell.

### Deferred error handling

Always check errors from deferred calls:
```go
// ❌ WRONG - linter will complain
defer file.Close()

// ✅ CORRECT - log the error
defer func() {
    if err := file.Close(); err != nil {
        logger.Error("failed to close file", "error", err)
    }
}()
```

For cleanup operations, logging is usually sufficient. For operations where failure matters, use named returns.

### Constructor signatures

Each implementation can have different constructor signatures:
```go
// Different constructors, same interface
func postgres.New(cfg *config.Config) (*Storage, error)
func local.New(cfg *config.Config) (*Storage, error)
```

The calling code (lab.NewManager) switches between them based on config.

## Testing approach

The codebase uses multiple testing strategies:

**Unit tests with mocks:**
- Provider mocks in `internal/provider/mock/`
- Lab manager mocks in `internal/lab/mock/`
- SSH mocks in `internal/util/serverchecker/mock_ssh.go`
- Example test structure in `cmd/create_lab_test.go` and `cmd/delete_lab_test.go`

**Integration tests:**
- PostgreSQL storage tests in `internal/storage/postgres/postgres_test.go`
- Tests against real PostgreSQL database (not mocked)
- Comprehensive coverage: Save, Get, List, Delete, error cases
- Template for future integration tests

**Running tests:**
```bash
# All tests
go test ./...

# Specific package
go test ./internal/storage/postgres/ -v

# With coverage
go test ./internal/storage/postgres/ -v -cover

# Specific test
go test ./internal/storage/postgres/ -v -run TestSaveAndGet
```

**Test structure template** (see `postgres_test.go`):
```go
func TestFeature(t *testing.T) {
    // Setup
    cfg := getTestConfig()
    storage, err := New(cfg)
    if err != nil {
        t.Fatalf("failed to create: %v", err)
    }
    defer storage.Close()
    defer cleanupTestData(t, storage, "test-name")

    // Test
    err = storage.Operation()
    if err != nil {
        t.Fatalf("operation failed: %v", err)
    }

    // Verify
    if result != expected {
        t.Errorf("expected %v, got %v", expected, result)
    }
    t.Logf("✓ Test passed")
}
```

**PostgreSQL tests require:**
- PostgreSQL running: `./scripts/start-postgres.sh`
- Migrations applied (script does this automatically)
- Connection to localhost:5432
