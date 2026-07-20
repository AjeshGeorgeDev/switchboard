#!/bin/sh
set -eu

MIGRATE_ON_START="${MIGRATE_ON_START:-true}"
MIGRATIONS_PATH="${MIGRATIONS_PATH:-/migrations}"

if [ "$MIGRATE_ON_START" = "true" ]; then
  if [ -z "${DATABASE_URL:-}" ]; then
    echo "DATABASE_URL is required when MIGRATE_ON_START=true" >&2
    exit 1
  fi
  echo "Running database migrations..."
  migrate -path "$MIGRATIONS_PATH" -database "$DATABASE_URL" up
  echo "Migrations complete."
fi

exec /server
