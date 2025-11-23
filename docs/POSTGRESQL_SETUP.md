# PostgreSQL setup for storctl development

This guide covers setting up PostgreSQL for local storctl development on macOS.

## Quick start

```bash
# Start PostgreSQL (creates container if needed, applies migrations)
./scripts/start-postgres.sh

# Run tests
go test ./internal/storage/postgres/ -v

# Stop PostgreSQL
./scripts/stop-postgres.sh
```

## Architecture

storctl uses PostgreSQL for persistent storage of lab metadata. The implementation supports both PostgreSQL (for production/development) and BoltDB (for local-only usage).

### Storage abstraction

```
storage.Storage (interface)
├── postgres.Storage (PostgreSQL implementation)
└── local.Storage (BoltDB implementation)
```

The storage backend is selected via `config.yaml`:

```yaml
storage:
  type: postgres  # or "local" for BoltDB
  postgres:
    host: localhost
    port: 5432
    database: storctl_dev
    user: postgres
    password: storctl
```

## Initial setup

### Prerequisites

- Podman installed (`brew install podman`)
- Podman machine initialized and running

### Start PostgreSQL

```bash
./scripts/start-postgres.sh
```

This script will:
1. Create data directory at `~/.storctl/postgres-data`
2. Start PostgreSQL container (or create if doesn't exist)
3. Wait for PostgreSQL to be ready
4. Apply migrations if database is empty

### Verify setup

```bash
# Check container is running
podman ps | grep storctl-postgres

# Connect to database
podman exec -it storctl-postgres psql -U postgres -d storctl_dev

# List tables
\dt

# Exit
\q
```

## Data persistence

### Where is data stored?

PostgreSQL data is stored in **`~/.storctl/postgres-data`** on your Mac filesystem (not inside the Podman VM).

**This means data survives:**
- ✅ Container restarts
- ✅ Container recreation
- ✅ Podman machine recreation
- ✅ macOS upgrades
- ❌ macOS reinstall (backup your home directory!)

### Why host mount instead of volume?

Podman volumes live inside the Podman VM. If you recreate the Podman machine with `podman machine rm` / `podman machine init`, all volumes are lost.

Using a host mount (`-v ~/.storctl/postgres-data:/var/lib/postgresql/data:Z`) stores data on your Mac, making it more durable.

### Backup and restore

**Create backup:**

```bash
./scripts/backup-postgres.sh
```

Backups are stored in `~/.storctl/backups/postgres-YYYYMMDD-HHMMSS.sql`. The script keeps the last 10 backups automatically.

**Restore from backup:**

```bash
./scripts/restore-postgres.sh ~/.storctl/backups/postgres-20250122-143022.sql
```

**Manual backup:**

```bash
podman exec storctl-postgres pg_dump -U postgres storctl_dev > backup.sql
```

**Manual restore:**

```bash
podman exec -i storctl-postgres psql -U postgres storctl_dev < backup.sql
```

## Migrations

Migrations are stored in `migrations/` directory:

- `001_initial.sql` - Creates `labs` and `audit_logs` tables

### Apply migrations manually

```bash
podman exec -i storctl-postgres psql -U postgres -d storctl_dev < migrations/001_initial.sql
```

### Create new migration

1. Create `migrations/002_description.sql`
2. Add your SQL changes
3. Apply manually (we'll add migration tooling later)

## Connection details

### From Go code

```go
cfg := &config.Config{
    Storage: config.StorageConfig{
        Type: "postgres",
        Postgres: config.PostgresConfig{
            Host:     "localhost",
            Port:     "5432",
            Database: "storctl_dev",
            User:     "postgres",
            Password: "storctl",
        },
    },
}

storage, err := postgres.New(cfg)
```

### Connection string

```
postgresql://postgres:storctl@localhost:5432/storctl_dev?sslmode=disable
```

**Note:** `sslmode=disable` is required for local development. Production should use SSL.

### From psql command line

```bash
psql -h localhost -U postgres -d storctl_dev
# Password: storctl
```

### From container

```bash
podman exec -it storctl-postgres psql -U postgres -d storctl_dev
```

## Useful commands

### Container management

```bash
# Start container
./scripts/start-postgres.sh

# Stop container
./scripts/stop-postgres.sh

# View logs
podman logs storctl-postgres

# Follow logs
podman logs -f storctl-postgres

# Restart container
podman restart storctl-postgres

# Remove container (data persists!)
podman stop storctl-postgres
podman rm storctl-postgres
```

### Database queries

```bash
# List all labs
podman exec -it storctl-postgres psql -U postgres -d storctl_dev -c "SELECT name, created_at FROM labs;"

# Count labs
podman exec -it storctl-postgres psql -U postgres -d storctl_dev -c "SELECT COUNT(*) FROM labs WHERE deleted_at IS NULL;"

# View deleted labs
podman exec -it storctl-postgres psql -U postgres -d storctl_dev -c "SELECT name, deleted_at FROM labs WHERE deleted_at IS NOT NULL;"

# Delete all test data
podman exec -it storctl-postgres psql -U postgres -d storctl_dev -c "DELETE FROM labs WHERE name LIKE 'test-%';"
```

### Database admin

```bash
# Database size
podman exec -it storctl-postgres psql -U postgres -d storctl_dev -c "SELECT pg_size_pretty(pg_database_size('storctl_dev'));"

# Table sizes
podman exec -it storctl-postgres psql -U postgres -d storctl_dev -c "SELECT tablename, pg_size_pretty(pg_total_relation_size(tablename::text)) FROM pg_tables WHERE schemaname = 'public';"

# Active connections
podman exec -it storctl-postgres psql -U postgres -d storctl_dev -c "SELECT * FROM pg_stat_activity WHERE datname = 'storctl_dev';"
```

## Testing

### Run storage tests

```bash
# All PostgreSQL tests
go test ./internal/storage/postgres/ -v

# Specific test
go test ./internal/storage/postgres/ -v -run TestSaveAndGet

# With coverage
go test ./internal/storage/postgres/ -v -cover
```

### Test with real lab creation

Update `~/.storctl/config.yaml`:

```yaml
storage:
  type: postgres
  postgres:
    host: localhost
    port: 5432
    database: storctl_dev
    user: postgres
    password: storctl
```

Then create a lab (with Lima for safety):

```bash
# Using existing config
storctl create lab test-postgres -f examples/lima-simple.yaml

# Verify in database
podman exec -it storctl-postgres psql -U postgres -d storctl_dev -c "SELECT name, status->>'state' AS state FROM labs;"

# Clean up
storctl delete lab test-postgres --force
```

## Troubleshooting

### Container won't start

```bash
# Check if port 5432 is already in use
lsof -i :5432

# Check Podman machine status
podman machine list
podman machine start

# Check container logs
podman logs storctl-postgres
```

### Connection refused

```bash
# Verify container is running
podman ps | grep storctl-postgres

# Check if PostgreSQL is ready
podman exec storctl-postgres pg_isready -U postgres

# Restart container
podman restart storctl-postgres
```

### SSL error: "SSL is not enabled on the server"

The PostgreSQL driver requires `sslmode=disable` for local development. This is already configured in `postgres.go`:

```go
connStr := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=disable", ...)
```

### Migration not applied

```bash
# Check if tables exist
podman exec -it storctl-postgres psql -U postgres -d storctl_dev -c "\dt"

# If missing, apply manually
podman exec -i storctl-postgres psql -U postgres -d storctl_dev < migrations/001_initial.sql
```

### Lost data after Podman machine recreation

This shouldn't happen if using the updated `start-postgres.sh` script (which uses host mounts). If you're using the old script with volumes:

1. Backup before recreating machine: `./scripts/backup-postgres.sh`
2. Recreate machine
3. Restore: `./scripts/restore-postgres.sh ~/.storctl/backups/postgres-XXXXXX.sql`

Better solution: Use the updated script with host mounts.

## Production considerations

For production deployment (not local development):

1. **Enable SSL**: Remove `sslmode=disable`, configure PostgreSQL with SSL certificates
2. **Use connection pooling**: Consider `pgbouncer` or similar
3. **Strong password**: Don't use `storctl` as password!
4. **Environment variables**: Don't commit passwords to config
5. **Backups**: Set up automated backups (e.g., to S3)
6. **Monitoring**: Set up metrics and alerts
7. **High availability**: Consider replicas, failover

## Environment variables

Override defaults with environment variables:

```bash
# Use different database
export POSTGRES_DB=storctl_test
./scripts/start-postgres.sh

# Use different password
export POSTGRES_PASSWORD=my-secure-password
./scripts/start-postgres.sh

# Use different version
export POSTGRES_VERSION=16
./scripts/start-postgres.sh
```

## Next steps

After PostgreSQL is working:

1. Implement BoltDB Get/List/Delete methods (for backward compatibility)
2. Add tests for BoltDB storage
3. Test end-to-end lab creation with PostgreSQL
4. Move to Week 2: Route53 DNS provider
