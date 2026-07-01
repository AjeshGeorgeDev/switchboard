-- name: ListApplications :many
SELECT * FROM applications ORDER BY sort_order, name;

-- name: ListPublicApplications :many
SELECT * FROM applications
WHERE is_active = TRUE AND is_public = TRUE
ORDER BY sort_order, name;

-- name: ListApplicationsForRoles :many
SELECT DISTINCT a.* FROM applications a
JOIN app_role_access ara ON ara.application_id = a.id
JOIN roles r ON r.id = ara.role_id
WHERE r.name = ANY($1::text[]) AND a.is_active = TRUE
ORDER BY a.sort_order, a.name;

-- name: GetApplicationByID :one
SELECT * FROM applications WHERE id = $1 LIMIT 1;

-- name: CreateApplication :one
INSERT INTO applications (name, description, icon_url, access_type, target_host, target_port, is_active, is_public, sort_order, section_id, created_by)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING *;

-- name: UpdateApplication :one
UPDATE applications SET
    name = $2, description = $3, icon_url = $4, access_type = $5,
    target_host = $6, target_port = $7, is_active = $8, is_public = $9, sort_order = $10,
    section_id = $11,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteApplication :exec
DELETE FROM applications WHERE id = $1;

-- name: GetApplicationRoles :many
SELECT r.* FROM roles r
JOIN app_role_access ara ON ara.role_id = r.id
WHERE ara.application_id = $1;

-- name: SetApplicationRoles :exec
DELETE FROM app_role_access WHERE application_id = $1;

-- name: AddApplicationRole :exec
INSERT INTO app_role_access (application_id, role_id) VALUES ($1, $2)
ON CONFLICT DO NOTHING;
