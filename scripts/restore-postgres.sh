#!/bin/bash
set -e

CONTAINER_NAME="storctl-postgres"
POSTGRES_DB="${POSTGRES_DB:-storctl_dev}"

if [ -z "$1" ]; then
    echo "Usage: $0 <backup-file>"
    echo ""
    echo "Available backups:"
    ls -lh "$HOME/.storctl/backups"/postgres-*.sql 2>/dev/null || echo "  (no backups found)"
    exit 1
fi

BACKUP_FILE="$1"

if [ ! -f "$BACKUP_FILE" ]; then
    echo "Error: Backup file not found: $BACKUP_FILE"
    exit 1
fi

# Check if container is running
if ! podman ps --format "{{.Names}}" | grep -q "^${CONTAINER_NAME}$"; then
    echo "Error: PostgreSQL container is not running"
    echo "Start it with: ./scripts/start-postgres.sh"
    exit 1
fi

echo "Restoring PostgreSQL database..."
echo "  Database: $POSTGRES_DB"
echo "  From: $BACKUP_FILE"
echo ""
read -p "This will overwrite existing data. Continue? (y/N) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Restore cancelled"
    exit 1
fi

podman exec -i $CONTAINER_NAME psql -U postgres $POSTGRES_DB < "$BACKUP_FILE"

echo ""
echo "✓ Restore complete"
