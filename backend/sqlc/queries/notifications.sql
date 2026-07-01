-- name: ListNotificationsForUser :many
SELECT * FROM notifications WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2;

-- name: ListUnreadNotificationsForUser :many
SELECT * FROM notifications WHERE user_id = $1 AND read = FALSE ORDER BY created_at DESC LIMIT $2;

-- name: CountUnreadNotifications :one
SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND read = FALSE;

-- name: CreateNotification :one
INSERT INTO notifications (user_id, title, body, event_type, severity, link_url)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: MarkNotificationRead :exec
UPDATE notifications SET read = TRUE WHERE id = $1 AND user_id = $2;

-- name: MarkAllNotificationsRead :exec
UPDATE notifications SET read = TRUE WHERE user_id = $1 AND read = FALSE;

-- name: DeleteOldNotifications :exec
DELETE FROM notifications WHERE created_at < $1;

-- name: GetNotificationPreferences :many
SELECT * FROM notification_preferences WHERE user_id = $1;

-- name: UpsertNotificationPreference :exec
INSERT INTO notification_preferences (user_id, channel, event_type, enabled)
VALUES ($1, $2, $3, $4)
ON CONFLICT (user_id, channel, event_type) DO UPDATE SET enabled = EXCLUDED.enabled;
