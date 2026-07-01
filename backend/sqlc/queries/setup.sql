-- name: HasAdminUser :one
SELECT EXISTS (
    SELECT 1
    FROM user_roles ur
    JOIN roles r ON r.id = ur.role_id
    WHERE r.name = 'admin'
) AS has_admin;
