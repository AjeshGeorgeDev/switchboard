-- name: CreateInvitation :one
INSERT INTO user_invitations (email, username, display_name, role_ids, token_hash, invited_by, expires_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetInvitationByTokenHash :one
SELECT * FROM user_invitations
WHERE token_hash = $1 AND accepted_at IS NULL AND expires_at > now()
LIMIT 1;

-- name: ListPendingInvitations :many
SELECT * FROM user_invitations
WHERE accepted_at IS NULL AND expires_at > now()
ORDER BY created_at DESC;

-- name: MarkInvitationAccepted :exec
UPDATE user_invitations SET accepted_at = now() WHERE id = $1;

-- name: GetPendingInvitationByEmail :one
SELECT * FROM user_invitations
WHERE lower(email) = lower($1) AND accepted_at IS NULL AND expires_at > now()
LIMIT 1;

-- name: GetPendingInvitationByUsername :one
SELECT * FROM user_invitations
WHERE lower(username) = lower($1) AND accepted_at IS NULL AND expires_at > now()
LIMIT 1;
