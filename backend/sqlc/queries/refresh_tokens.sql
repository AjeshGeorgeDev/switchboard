-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (user_id, token_hash, expires_at, user_agent, ip_address)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetRefreshTokenByHash :one
SELECT * FROM refresh_tokens WHERE token_hash = $1 AND revoked = FALSE LIMIT 1;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens SET revoked = TRUE WHERE id = $1;

-- name: RevokeAllUserRefreshTokens :exec
UPDATE refresh_tokens SET revoked = TRUE WHERE user_id = $1 AND revoked = FALSE;

-- name: ListUserRefreshTokens :many
SELECT * FROM refresh_tokens WHERE user_id = $1 AND revoked = FALSE ORDER BY issued_at DESC;

-- name: ListUserLoginHistory :many
SELECT id, user_id, issued_at, expires_at, revoked, user_agent, ip_address
FROM refresh_tokens
WHERE user_id = $1
ORDER BY issued_at DESC
LIMIT 100;
