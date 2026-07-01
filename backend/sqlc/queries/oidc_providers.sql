-- name: ListOIDCProviders :many
SELECT * FROM oidc_providers ORDER BY name;

-- name: ListActiveOIDCProviders :many
SELECT id, name, display_name, issuer_url, client_id, scopes, auto_provision, default_role_id, is_active, created_at
FROM oidc_providers WHERE is_active = TRUE ORDER BY name;

-- name: GetOIDCProviderByName :one
SELECT * FROM oidc_providers WHERE name = $1 LIMIT 1;

-- name: GetOIDCProviderByID :one
SELECT * FROM oidc_providers WHERE id = $1 LIMIT 1;

-- name: CreateOIDCProvider :one
INSERT INTO oidc_providers (name, display_name, issuer_url, client_id, client_secret, scopes, auto_provision, default_role_id, is_active)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: UpdateOIDCProvider :one
UPDATE oidc_providers SET
    display_name = $2, issuer_url = $3, client_id = $4, client_secret = $5,
    scopes = $6, auto_provision = $7, default_role_id = $8, is_active = $9
WHERE id = $1
RETURNING *;

-- name: DeleteOIDCProvider :exec
DELETE FROM oidc_providers WHERE id = $1;
