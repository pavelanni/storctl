# Storage Abstraction Layer with PostgreSQL Support

## Status

Accepted

## Date

2025-11-22

## Context

- Original design used BoltDB exclusively for local lab metadata storage (ADR-0001)
- Need for remote state storage to solve "state fragmentation" problem (Terraform states scattered on developer laptops)
- Requirement for centralized visibility: all team members need to see all labs
- Need for cost tracking and TTL automation (prototype goals)
- Multiple developers creating labs without coordination leads to forgotten resources
- BoltDB is single-process, file-based, and cannot provide centralized access
- Production deployment will need robust, scalable database with ACID guarantees
- Need to maintain backward compatibility with existing BoltDB usage for local development

## Decision

Implement a storage abstraction layer with multiple backend implementations:

1. **Storage Interface**
   - Define common interface: `Save()`, `Get()`, `List()`, `Delete()`, `Close()`
   - Located in `internal/storage/interface.go`
   - Allows swapping storage backends via configuration

2. **PostgreSQL Implementation** (`internal/storage/postgres/`)
   - Primary backend for production and remote state scenarios
   - Features:
     - ACID transactions
     - Soft deletes (sets `deleted_at` timestamp)
     - UPSERT behavior (ON CONFLICT DO UPDATE)
     - Connection pooling support
     - Structured logging integration
   - Schema in `migrations/001_initial.sql`
   - Uses `github.com/lib/pq` driver
   - Connection string includes `sslmode=disable` for local dev (must enable SSL in production)

3. **BoltDB Implementation** (`internal/storage/local/`)
   - Backward compatibility for local-only usage
   - Embedded database, no server required
   - Maintains existing behavior for developers who don't need remote state

4. **Configuration-Based Selection**
   - Backend selected via `config.yaml`:
     ```yaml
     storage:
       type: postgres  # or "local"
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
   - Lab manager (`internal/lab/lab.go`) switches between implementations at initialization

5. **Development Tooling**
   - Helper scripts in `scripts/` for PostgreSQL management
   - `start-postgres.sh` uses host mounts for data persistence (survives Podman machine recreation)
   - Backup/restore scripts for data protection
   - Comprehensive test suite in `internal/storage/postgres/postgres_test.go`

## Consequences

### Positive

- **Centralized state**: All developers see the same lab state (no more fragmentation)
- **Cost visibility**: Database enables queries for cost tracking and reporting
- **TTL enforcement**: Daemon can query database for expired labs and auto-delete
- **ACID guarantees**: PostgreSQL transactions prevent corruption
- **Scalability**: PostgreSQL can handle production workloads
- **Audit trail**: Database supports audit_logs table for compliance
- **Backward compatibility**: Existing BoltDB usage still supported for local-only scenarios
- **Clean architecture**: Interface-based design makes future storage backends easy to add
- **Testability**: Integration tests verify PostgreSQL behavior without mocks

### Negative

- **Complexity**: Additional setup required (PostgreSQL container/server)
- **Dependency**: Requires PostgreSQL for remote state scenarios (but BoltDB still available for local)
- **Migration**: Need to migrate existing BoltDB data to PostgreSQL (manual process)
- **Maintenance**: Database requires backups, monitoring, updates
- **Development overhead**: Developers need to run PostgreSQL locally for testing
- **Network dependency**: Remote state requires network connectivity (BoltDB is offline-capable)

## Related Decisions

- [ADR-0001](0001-use-boltdb-as-local-storage.md) - Use of BBolt as Local Storage (still valid for local mode)
- [ADR-0009](0009-logging-strategy.md) - Logging Strategy (used in storage implementations)
- [ADR-0002](0002-cloud-provider-interface.md) - Cloud Provider Interface (similar abstraction pattern)

## Notes

- This decision enables the "Week 1" deliverable from the prototype plan (docs/planning/PROTOTYPE_PLAN.md)
- PostgreSQL chosen over alternatives (S3, Redis, etc.) for:
  - Strong consistency and ACID guarantees
  - Rich query capabilities for cost analysis
  - Native locking for TTL daemon coordination
  - Familiar to ops teams
  - Mature Go ecosystem support
- BoltDB implementation is intentionally kept as fallback for:
  - Local development without PostgreSQL
  - Offline scenarios (training labs without internet)
  - Simple single-user deployments
- Data persistence strategy for development: Host mounts (`~/.storctl/postgres-data`) instead of Podman volumes to survive VM recreation
- See `docs/POSTGRESQL_SETUP.md` for complete setup and usage documentation
- Future consideration: Add connection pooling (pgbouncer) for production deployments
