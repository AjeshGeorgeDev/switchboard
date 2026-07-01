#!/usr/bin/env bash
# Creates or updates a local-only user for README screenshot capture.
set -euo pipefail

HASH='$2a$10$d.FtcVzEbw4cPETb6s88I.G0MPd1MLpiIk2JhbqBe7a97DlS6DUbi'
CONTAINER="${POSTGRES_CONTAINER:-switchboard-postgres-1}"

docker exec -i "$CONTAINER" psql -U switchboard -d switchboard <<SQL
INSERT INTO users (username, email, display_name, auth_type, password_hash)
VALUES ('screenshot', 'screenshot@switchboard.local', 'Demo User', 'local', '${HASH}')
ON CONFLICT (email) DO UPDATE SET
  password_hash = EXCLUDED.password_hash,
  is_active = TRUE;

INSERT INTO user_roles (user_id, role_id)
SELECT u.id, r.id
FROM users u
JOIN roles r ON r.name = 'admin'
WHERE u.email = 'screenshot@switchboard.local'
ON CONFLICT DO NOTHING;
SQL

echo "Screenshot user ready: screenshot@switchboard.local / screenshot-demo"
