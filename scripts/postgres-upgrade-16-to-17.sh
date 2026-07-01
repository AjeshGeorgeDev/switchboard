#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BACKUP_DIR="${ROOT}/backups"
BACKUP_FILE="${BACKUP_DIR}/switchboard-pre-pg17.sql"
COMPOSE=(docker compose -f "${ROOT}/docker-compose.yml")
COMPOSE_PG16=(docker compose -f "${ROOT}/docker-compose.yml" -f "${ROOT}/docker-compose.pg16-backup.yml")
VOLUME_NAME="${COMPOSE_PROJECT_NAME:-switchboard}_postgres_data"

cd "${ROOT}"
mkdir -p "${BACKUP_DIR}"

echo "==> Starting Postgres 16 (reads your existing data volume)..."
"${COMPOSE_PG16[@]}" up -d postgres

echo "==> Waiting for Postgres 16..."
until "${COMPOSE_PG16[@]}" exec -T postgres pg_isready -U switchboard >/dev/null 2>&1; do
  sleep 1
done

echo "==> Dumping database to ${BACKUP_FILE}"
"${COMPOSE_PG16[@]}" exec -T postgres pg_dump -U switchboard --clean --if-exists switchboard > "${BACKUP_FILE}"

echo "==> Stopping containers..."
"${COMPOSE[@]}" down

echo "==> Removing old PG16 data volume (your data is saved in ${BACKUP_FILE})"
docker volume rm "${VOLUME_NAME}"

echo "==> Starting Postgres 17 with a fresh data directory..."
"${COMPOSE[@]}" up -d postgres

echo "==> Waiting for Postgres 17..."
until "${COMPOSE[@]}" exec -T postgres pg_isready -U switchboard >/dev/null 2>&1; do
  sleep 1
done

echo "==> Restoring backup into Postgres 17..."
"${COMPOSE[@]}" exec -T postgres psql -U switchboard -d switchboard -v ON_ERROR_STOP=1 < "${BACKUP_FILE}"

echo "==> Starting Redis..."
"${COMPOSE[@]}" up -d redis

echo ""
echo "Upgrade complete. Backup kept at:"
echo "  ${BACKUP_FILE}"
echo ""
echo "Verify with: docker compose ps && docker compose logs postgres"
