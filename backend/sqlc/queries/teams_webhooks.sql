-- name: ListTeamsWebhookConfigs :many
SELECT * FROM teams_webhook_configs ORDER BY name;

-- name: GetTeamsWebhookConfigByID :one
SELECT * FROM teams_webhook_configs WHERE id = $1 LIMIT 1;

-- name: ListActiveTeamsWebhookConfigs :many
SELECT * FROM teams_webhook_configs WHERE is_active = TRUE;

-- name: CreateTeamsWebhookConfig :one
INSERT INTO teams_webhook_configs (name, webhook_url, event_types, is_active)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateTeamsWebhookConfig :one
UPDATE teams_webhook_configs SET name = $2, webhook_url = $3, event_types = $4, is_active = $5
WHERE id = $1
RETURNING *;

-- name: DeleteTeamsWebhookConfig :exec
DELETE FROM teams_webhook_configs WHERE id = $1;
