-- name: CreateWebhookEvent :one
INSERT INTO webhook_events (source, status, payload, payload_preview)
VALUES ($1, 'accepted', $2, $3)
RETURNING *;

-- name: UpdateWebhookEventStatus :exec
UPDATE webhook_events
SET status = $2, error_message = $3, processed_at = now()
WHERE id = $1;

-- name: ListWebhookEvents :many
SELECT * FROM webhook_events
WHERE ($1 = '' OR source::text = $1)
ORDER BY received_at DESC
LIMIT $2 OFFSET $3;

-- name: CountWebhookEvents :one
SELECT COUNT(*) FROM webhook_events
WHERE ($1 = '' OR source::text = $1);

-- name: GetWebhookEventByID :one
SELECT * FROM webhook_events WHERE id = $1 LIMIT 1;

-- name: DeleteOldWebhookEvents :exec
DELETE FROM webhook_events WHERE received_at < $1;
