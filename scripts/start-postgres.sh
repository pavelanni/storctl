#!/bin/bash
set -e

DATA_DIR="$HOME/.storctl/postgres-data"
CONTAINER_NAME="storctl-postgres"
POSTGRES_PASSWORD="${POSTGRES_PASSWORD:-storctl}"
POSTGRES_DB="${POSTGRES_DB:-storctl_dev}"
POSTGRES_VERSION="${POSTGRES_VERSION:-17}"

echo "Starting PostgreSQL for storctl development..."

# Create data directory if it doesn't exist
mkdir -p "$DATA_DIR"

# Check if container exists and is running
if podman ps -a --format "{{.Names}}" | grep -q "^${CONTAINER_NAME}$"; then
    if podman ps --format "{{.Names}}" | grep -q "^${CONTAINER_NAME}$"; then
        echo "✓ PostgreSQL is already running"
        exit 0
    else
        echo "Starting existing container..."
        podman start $CONTAINER_NAME
        echo "✓ PostgreSQL started"
        exit 0
    fi
fi

# Create and start new container with host mount
echo "Creating new PostgreSQL container..."
podman run -d \
  --name $CONTAINER_NAME \
  -e POSTGRES_PASSWORD=$POSTGRES_PASSWORD \
  -e POSTGRES_DB=$POSTGRES_DB \
  -p 5432:5432 \
  -v ${DATA_DIR}:/var/lib/postgresql/data:Z \
  postgres:$POSTGRES_VERSION

echo "Waiting for PostgreSQL to be ready..."
for i in {1..30}; do
    if podman exec $CONTAINER_NAME pg_isready -U postgres > /dev/null 2>&1; then
        break
    fi
    sleep 1
    echo -n "."
done
echo ""

# Check if migrations are needed
echo "Checking database schema..."
if ! podman exec $CONTAINER_NAME psql -U postgres -d $POSTGRES_DB -c "\dt" 2>/dev/null | grep -q "labs"; then
    echo "Applying migrations..."
    if [ -f "migrations/001_initial.sql" ]; then
        podman exec -i $CONTAINER_NAME psql -U postgres -d $POSTGRES_DB < migrations/001_initial.sql
        echo "✓ Migrations applied"
    else
        echo "⚠ Warning: migrations/001_initial.sql not found"
        echo "  Run migrations manually: podman exec -i $CONTAINER_NAME psql -U postgres -d $POSTGRES_DB < migrations/001_initial.sql"
    fi
else
    echo "✓ Database already initialized"
fi

echo ""
echo "✓ PostgreSQL is ready!"
echo "  Container: $CONTAINER_NAME"
echo "  Database: $POSTGRES_DB"
echo "  Data location: $DATA_DIR"
echo "  Connection: postgresql://postgres:$POSTGRES_PASSWORD@localhost:5432/$POSTGRES_DB?sslmode=disable"
echo ""
echo "Useful commands:"
echo "  Stop:    podman stop $CONTAINER_NAME"
echo "  Start:   podman start $CONTAINER_NAME"
echo "  Logs:    podman logs $CONTAINER_NAME"
echo "  Shell:   podman exec -it $CONTAINER_NAME psql -U postgres -d $POSTGRES_DB"
echo "  Backup:  ./scripts/backup-postgres.sh"
