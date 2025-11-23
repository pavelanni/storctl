#!/bin/bash
set -e

BACKUP_DIR="$HOME/.storctl/backups"
CONTAINER_NAME="storctl-postgres"
POSTGRES_DB="${POSTGRES_DB:-storctl_dev}"
BACKUP_FILE="$BACKUP_DIR/postgres-$(date +%Y%m%d-%H%M%S).sql"

# Check if container is running
if ! podman ps --format "{{.Names}}" | grep -q "^${CONTAINER_NAME}$"; then
    echo "Error: PostgreSQL container is not running"
    echo "Start it with: ./scripts/start-postgres.sh"
    exit 1
fi

mkdir -p "$BACKUP_DIR"

echo "Backing up PostgreSQL database..."
echo "  Database: $POSTGRES_DB"
echo "  Output: $BACKUP_FILE"

podman exec $CONTAINER_NAME pg_dump -U postgres $POSTGRES_DB > "$BACKUP_FILE"

# Keep only last 10 backups
echo "Cleaning old backups (keeping last 10)..."
ls -t "$BACKUP_DIR"/postgres-*.sql 2>/dev/null | tail -n +11 | xargs rm -f 2>/dev/null || true

echo ""
echo "✓ Backup complete: $BACKUP_FILE"
echo ""
echo "Available backups:"
ls -lh "$BACKUP_DIR"/postgres-*.sql 2>/dev/null || echo "  (no backups found)"
echo ""
echo "To restore: ./scripts/restore-postgres.sh $BACKUP_FILE"
