-- name: CreateEmailOutboundLog :one
INSERT INTO email_outbound_log (event_type, subject, body_preview, status, error_message, triggered_by)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: CreateEmailOutboundRecipient :one
INSERT INTO email_outbound_recipients (log_id, email, user_id, status, error_message)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ListEmailOutboundLog :many
SELECT * FROM email_outbound_log
WHERE ($1 = '' OR event_type = $1)
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountEmailOutboundLog :one
SELECT COUNT(*) FROM email_outbound_log
WHERE ($1 = '' OR event_type = $1);

-- name: ListEmailOutboundRecipientsByLogIDs :many
SELECT * FROM email_outbound_recipients
WHERE log_id = ANY($1::uuid[])
ORDER BY email;

-- name: GetEmailOutboundLogByID :one
SELECT * FROM email_outbound_log WHERE id = $1 LIMIT 1;
