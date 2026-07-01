-- name: CreateAuditLog :one
INSERT INTO audit_logs (actor_id, actor_username, action, resource_type, resource_id, details, ip_address)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: ListAuditLogs :many
SELECT * FROM audit_logs
WHERE ($1::text IS NULL OR $1 = '' OR action = $1)
  AND ($2::text IS NULL OR $2 = '' OR resource_type = $2)
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: CountAuditLogs :one
SELECT COUNT(*) FROM audit_logs
WHERE ($1::text IS NULL OR $1 = '' OR action = $1)
  AND ($2::text IS NULL OR $2 = '' OR resource_type = $2);
