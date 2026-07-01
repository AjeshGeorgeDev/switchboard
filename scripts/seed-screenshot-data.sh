#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${APP_BASE_URL:-http://localhost:8080}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

post_harbor() {
  local file="$1"
  curl -sf -X POST "${BASE_URL}/webhooks/harbor" \
    -H "Content-Type: application/json" \
    --data-binary @"${file}" >/dev/null
  echo "  posted $(basename "$file")"
}

post_trivy() {
  local file="$1"
  curl -sf -X POST "${BASE_URL}/webhooks/trivy" \
    -H "Content-Type: application/json" \
    --data-binary @"${file}" >/dev/null
  echo "  posted $(basename "$file")"
}

echo "Seeding Harbor deployment reports..."
for f in "${SCRIPT_DIR}"/screenshot-data/harbor-*.json; do
  post_harbor "$f"
done

echo "Seeding Trivy CVE findings..."
for f in "${SCRIPT_DIR}"/screenshot-data/trivy-*.json; do
  post_trivy "$f"
done

echo "Waiting for async job processing..."
sleep 3
echo "Done."
