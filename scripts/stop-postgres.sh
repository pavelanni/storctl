#!/bin/bash
set -e

CONTAINER_NAME="storctl-postgres"

if podman ps --format "{{.Names}}" | grep -q "^${CONTAINER_NAME}$"; then
    echo "Stopping PostgreSQL container..."
    podman stop $CONTAINER_NAME
    echo "✓ PostgreSQL stopped"
else
    echo "PostgreSQL container is not running"
fi
