-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1 LIMIT 1;

-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = $1 LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1 LIMIT 1;

-- name: GetUserByOIDC :one
SELECT * FROM users WHERE oidc_provider = $1 AND oidc_subject = $2 LIMIT 1;

-- name: ListUsers :many
SELECT * FROM users ORDER BY created_at DESC;

-- name: CreateUser :one
INSERT INTO users (username, email, display_name, auth_type, password_hash, oidc_provider, oidc_subject, oidc_email)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: UpdateUser :one
UPDATE users SET
    email = COALESCE(sqlc.narg('email'), email),
    display_name = COALESCE(sqlc.narg('display_name'), display_name),
    auth_type = COALESCE(sqlc.narg('auth_type'), auth_type),
    password_hash = COALESCE(sqlc.narg('password_hash'), password_hash),
    oidc_provider = COALESCE(sqlc.narg('oidc_provider'), oidc_provider),
    oidc_subject = COALESCE(sqlc.narg('oidc_subject'), oidc_subject),
    oidc_email = COALESCE(sqlc.narg('oidc_email'), oidc_email),
    is_active = COALESCE(sqlc.narg('is_active'), is_active),
    last_login_at = COALESCE(sqlc.narg('last_login_at'), last_login_at),
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: UpdateUserLastLogin :exec
UPDATE users SET last_login_at = now(), updated_at = now() WHERE id = $1;

-- name: GetUserRoles :many
SELECT r.* FROM roles r
JOIN user_roles ur ON ur.role_id = r.id
WHERE ur.user_id = $1;

-- name: SetUserRoles :exec
DELETE FROM user_roles WHERE user_id = $1;

-- name: AddUserRole :exec
INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: GetUsersByRoleName :many
SELECT u.* FROM users u
JOIN user_roles ur ON ur.user_id = u.id
JOIN roles r ON r.id = ur.role_id
WHERE r.name = $1 AND u.is_active = TRUE;
