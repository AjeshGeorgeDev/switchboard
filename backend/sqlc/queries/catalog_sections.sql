-- name: ListCatalogSections :many
SELECT * FROM catalog_sections ORDER BY sort_order, name;

-- name: GetCatalogSectionByID :one
SELECT * FROM catalog_sections WHERE id = $1 LIMIT 1;

-- name: CreateCatalogSection :one
INSERT INTO catalog_sections (name, sort_order)
VALUES ($1, $2)
RETURNING *;

-- name: UpdateCatalogSection :one
UPDATE catalog_sections SET
    name = $2,
    sort_order = $3
WHERE id = $1
RETURNING *;

-- name: DeleteCatalogSection :exec
DELETE FROM catalog_sections WHERE id = $1;
