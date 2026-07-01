-- name: ListDeploymentReports :many
SELECT * FROM deployment_reports ORDER BY received_at DESC LIMIT $1 OFFSET $2;

-- name: ListDeploymentReportsFiltered :many
SELECT * FROM deployment_reports
WHERE ($1 = '' OR app_name ILIKE '%' || $1 || '%' OR image_name ILIKE '%' || $1 || '%')
  AND ($2 = '' OR status::text = $2)
ORDER BY received_at DESC
LIMIT $3 OFFSET $4;

-- name: CountDeploymentReports :one
SELECT COUNT(*) FROM deployment_reports;

-- name: CountDeploymentReportsFiltered :one
SELECT COUNT(*) FROM deployment_reports
WHERE ($1 = '' OR app_name ILIKE '%' || $1 || '%' OR image_name ILIKE '%' || $1 || '%')
  AND ($2 = '' OR status::text = $2);

-- name: GetDeploymentReportByID :one
SELECT * FROM deployment_reports WHERE id = $1 LIMIT 1;

-- name: GetDeploymentReportByPayloadHash :one
SELECT * FROM deployment_reports WHERE payload_hash = $1 LIMIT 1;

-- name: CreateDeploymentReport :one
INSERT INTO deployment_reports (app_name, image_name, image_tag, triggered_by, status, critical_count, high_count, medium_count, low_count, report_url, raw_payload, payload_hash)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
RETURNING *;

-- name: DeleteOldDeploymentReports :exec
DELETE FROM deployment_reports WHERE received_at < $1;
