#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BACKUP_DIR="${ROOT}/backups"
STAMP="$(date +%Y%m%d-%H%M%S)"
BACKUP_FILE="${BACKUP_DIR}/switchboard-${STAMP}.sql"
COMPOSE=(docker compose -f "${ROOT}/docker-compose.yml")

cd "${ROOT}"
mkdir -p "${BACKUP_DIR}"

"${COMPOSE[@]}" up -d postgres
until "${COMPOSE[@]}" exec -T postgres pg_isready -U switchboard >/dev/null 2>&1; do
  sleep 1
done

"${COMPOSE[@]}" exec -T postgres pg_dump -U switchboard --clean --if-exists switchboard > "${BACKUP_FILE}"
echo "Backup written to ${BACKUP_FILE}"
