# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project overview

storctl is a CLI tool for managing demo and lab environments for MinIO AIStor testing, training, and demonstrations on Hetzner Cloud. The tool manages the complete lifecycle of lab environments including servers, volumes, DNS records, SSH keys, and automated Kubernetes + AIStor installation via Ansible. All lab state is stored in a centralized PostgreSQL database for multi-user access and persistence.

## Architecture

### Core concepts

The codebase follows a provider pattern where cloud infrastructure providers (currently Hetzner Cloud, with future support for AWS and DigitalOcean planned) implement a common `CloudProvider` interface. Resources follow a Kubernetes-style API model with TypeMeta, ObjectMeta, Spec, and Status fields. The provider interface enables easy addition of new cloud providers without changing core logic.

### Key packages

- `cmd/` - Cobra CLI commands organized by action (create, get, delete, install)
- `internal/types` - Core resource types (Lab, Server, Volume, SSHKey) with Kubernetes-style structure
- `internal/provider` - Provider interface and implementations
  - `provider/hetzner` - Hetzner Cloud provider implementation (production)
  - `provider/mock` - Mock provider for testing
  - `provider/factory.go` - Factory function to create providers based on config
- `internal/storage` - Storage abstraction layer
  - `storage/interface.go` - Storage interface (Save, Get, List, Delete, Close)
  - `storage/postgres/` - PostgreSQL implementation (only backend supported)
- `internal/lab` - Lab manager that orchestrates multi-resource operations
  - Uses PostgreSQL storage for centralized state
  - Ansible inventory generation and playbook execution
  - Provider-specific lab creation/deletion logic
- `internal/config` - Configuration management and constants
- `internal/ssh` - SSH key management
- `internal/dns` - DNS management (currently Cloudflare, Route53 planned)
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

1. Lab creation (`Create`) creates servers, volumes, and SSH keys via Hetzner Cloud
1. Creates lab-specific SSH key for Ansible access
1. Creates servers and waits for SSH readiness (using `serverchecker` package)
1. Creates volumes and attaches to servers
1. **CRITICAL:** Provider-specific creation functions **must** populate `lab.Status`:
   - `lab.Status.Servers = servers` after server creation
   - `lab.Status.Volumes = volumes` after volume creation
   - Without this, downstream operations (DNS, Ansible) will fail
1. Lab data stored in PostgreSQL database
   - Centralized remote database with soft deletes
   - ACID transactions ensure data consistency
   - Supports multi-user access and persistence
1. Lab listing supports `--show-deleted` flag to view soft-deleted labs
1. `SyncLabs()` fetches labs from provider by querying servers with `lab_name` label
1. `install lab` generates Ansible inventory and runs embedded playbooks to install K3s + AIStor

## Development commands

### Common CLI operations

```bash
# List active labs from storage
go run . get lab

# List all labs including soft-deleted (PostgreSQL only)
go run . get lab --show-deleted

# Get specific lab details
go run . get lab <lab-name>

# Create lab from manifest
go run . create -f examples/lab-hetzner-snsd.yaml

# Delete lab (soft delete in PostgreSQL)
go run . delete lab <lab-name>
```

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

- `config.yaml` - Main config (Hetzner credentials, DNS, PostgreSQL connection)
- `postgres-data/` - PostgreSQL data directory (when using local PostgreSQL with host mount)
- `backups/` - PostgreSQL database backups (created by `scripts/backup-postgres.sh`)
- `keys/` - SSH keys for lab access
- `ansible/` - Generated Ansible inventory files
- `templates/` - Lab templates (YAML manifests)

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

### Optional provider for storage operations

The lab manager supports `nil` provider for storage-only operations:

```go
// Storage-only operations (List, Get from storage)
labSvc, err := lab.NewManager(nil, cfg)  // provider = nil
labs, err := labSvc.List(false)          // works without provider

// Operations requiring provider will return clear error
err := labSvc.Create(lab)  // Error: "provider is required for Create operation"
```

**When provider is required:**
- `Create()` - Creates resources on cloud provider
- `Delete()` - Deletes resources from cloud provider
- `SyncLabs()` - Fetches labs from cloud provider

**When provider is optional (can be nil):**
- `List()` - Lists labs from storage backend
- `Get()` - Gets lab from storage (falls back to provider sync if not found)

This pattern improves performance by avoiding unnecessary provider initialization and API calls when only querying local/remote storage.

### Time-to-live (TTL)

Resources support TTL for automatic cleanup. TTL is parsed in `internal/util/timeutil` and stored as `DeleteAfter` timestamp in Status. Default TTL is 1 hour (see `config.DefaultTTL`).

### Label selectors

Resources use labels for filtering and grouping. Lab name is stored as `lab_name` label on all child resources. Use `LabelSelector` in ListOpts for querying (format: `"key=value"`).

### Server readiness checking

Hetzner labs use `internal/util/serverchecker` to verify SSH connectivity before proceeding with Ansible. This prevents race conditions where cloud-init hasn't finished.

### Ansible integration

- Playbooks embedded via `assets/PlaybookFiles` (Go embed.FS)
- Inventory generated dynamically from lab servers in JSON format
- Uses 'ansible' user on Hetzner Cloud servers (configured via cloud-init)
- Playbooks install K3s, DirectPV, Helm, and AIStor

### Storage abstraction

The storage layer uses an interface-based architecture supporting multiple backends:

```go
// internal/storage/interface.go
type Storage interface {
    Save(lab *types.Lab) error
    Get(name string) (*types.Lab, error)
    List(showDeleted bool) ([]*types.Lab, error)
    Delete(name string) error
    Close() error
}
```

**Implementations:**
- **PostgreSQL** (`internal/storage/postgres/`) - Remote database (only supported backend)
  - Full CRUD operations with soft deletes
  - `Delete()` sets `deleted_at` timestamp (soft delete)
  - `List(showDeleted)` conditionally filters deleted labs:
    - `List(false)` - Only active labs (WHERE deleted_at IS NULL)
    - `List(true)` - All labs including soft-deleted
  - UPSERT on conflict (ON CONFLICT DO UPDATE)
  - Requires `_ "github.com/lib/pq"` driver import
  - Connection string: `host=%s port=%s dbname=%s user=%s password=%s sslmode=disable`
  - Schema in `migrations/001_initial.sql` (labs + audit_logs tables)

**Note:** BoltDB support was removed in favor of centralized PostgreSQL for multi-user access. The Storage interface remains for potential future backends (e.g., Hetzner Object Storage).

**Configuration** (`config.yaml`):
```yaml
storage:
  type: postgres  # defaults to postgres if not specified
  postgres:
    host: localhost
    port: 5432
    database: storctl_dev
    user: postgres
    password: storctl
```

**Backend selection** in `internal/lab/lab.go`:
```go
func NewManager(provider provider.CloudProvider, cfg *config.Config) (*ManagerSvc, error) {
    var storage storage.Storage  // interface, NOT pointer to interface

    switch cfg.Storage.Type {
    case "postgres", "":  // empty string defaults to postgres
        storage, err = postgres.New(cfg)
        if err != nil {
            return nil, fmt.Errorf("failed to create lab storage: %w", err)
        }
    default:
        return nil, fmt.Errorf("invalid storage type: %s (only 'postgres' is supported)", cfg.Storage.Type)
    }

    return &ManagerSvc{
        Storage: storage,  // No &, interfaces are already reference types
    }
}
```

**Important:** Never use pointer to interface (`*storage.Storage`). Interfaces are already reference types internally.

## Known issues and solutions

### Empty lab.Status after Hetzner creation (FIXED)

**Problem:** Lab created successfully on Hetzner but downstream operations (DNS, Ansible) failed because `lab.Status.Servers` was empty.

**Root cause:** The `createLabHetzner()` function collected servers and volumes in local variables but never assigned them to `lab.Status` before returning.

**Solution:** Always assign collected resources to lab.Status:
```go
// After creating servers
servers := make([]*types.Server, 0)
for _, serverSpec := range specServers {
    result, err := m.Provider.CreateServer(...)
    servers = append(servers, result)
}
// ✅ CRITICAL: Assign to lab.Status
lab.Status.Servers = servers

// After creating volumes
createdVolumes := make([]*types.Volume, 0, len(volumes))
for _, volumeSpec := range volumes {
    volume, err := m.Provider.CreateVolume(...)
    createdVolumes = append(createdVolumes, volume)
}
// ✅ CRITICAL: Assign to lab.Status
lab.Status.Volumes = createdVolumes
```

**Prevention:** When adding new provider implementations, ensure all Status fields are properly populated before returning from creation functions.

### PostgreSQL List() returning empty results (FIXED)

**Problem:** Labs visible in database but `List()` returned empty array with no error.

**Root cause:** Missing `rows.Err()` check after `rows.Next()` loop. In Go's `database/sql`, iteration errors don't surface until you call `rows.Err()`.

**Solution:** Always check `rows.Err()` after iteration:
```go
for rows.Next() {
    // scan rows...
}
// ✅ Check for iteration errors
if err = rows.Err(); err != nil {
    return nil, fmt.Errorf("error iterating over rows: %w", err)
}
```

### Index out of range panic in DNS creation (FIXED)

**Problem:** Panic when accessing `lab.Status.Servers[0]` during DNS record creation.

**Root cause:** Code assumed servers exist without checking. Could happen if Status wasn't populated or if lab has no servers.

**Solution:** Always check slice length before accessing:
```go
// ✅ Check before accessing
if len(lab.Status.Servers) == 0 {
    return fmt.Errorf("no servers found in lab status, cannot create DNS records")
}
cpPublicNet := lab.Status.Servers[0].Status.PublicNet
```

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
func hetzner.New(cfg *config.Config) (CloudProvider, error)
```

The calling code (factory functions) switches between them based on config. Each implementation returns the interface type, keeping the calling code generic.

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
