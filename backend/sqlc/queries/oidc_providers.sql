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
INSERT INTO oidc_providers (
    name, display_name, issuer_url, client_id, client_secret, scopes,
    auto_provision, default_role_id, is_active,
    claim_email, claim_name, claim_subject, claim_groups, group_role_mappings
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
RETURNING *;

-- name: UpdateOIDCProvider :one
UPDATE oidc_providers SET
    display_name = sqlc.arg(display_name),
    issuer_url = sqlc.arg(issuer_url),
    client_id = sqlc.arg(client_id),
    client_secret = COALESCE(NULLIF(sqlc.arg(client_secret)::text, ''), oidc_providers.client_secret),
    scopes = sqlc.arg(scopes),
    auto_provision = sqlc.arg(auto_provision),
    default_role_id = sqlc.arg(default_role_id),
    is_active = sqlc.arg(is_active),
    claim_email = sqlc.arg(claim_email),
    claim_name = sqlc.arg(claim_name),
    claim_subject = sqlc.arg(claim_subject),
    claim_groups = sqlc.arg(claim_groups),
    group_role_mappings = sqlc.arg(group_role_mappings)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteOIDCProvider :exec
DELETE FROM oidc_providers WHERE id = $1;
