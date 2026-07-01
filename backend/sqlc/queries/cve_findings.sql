-- name: ListCVEFindings :many
SELECT * FROM cve_findings ORDER BY scan_date DESC, severity LIMIT $1 OFFSET $2;

-- name: ListCVEFindingsFiltered :many
SELECT * FROM cve_findings
WHERE ($1 = '' OR severity::text = $1)
  AND ($2 = '' OR image_name ILIKE '%' || $2 || '%' OR cve_id ILIKE '%' || $2 || '%')
ORDER BY scan_date DESC, severity
LIMIT $3 OFFSET $4;

-- name: CountCVEFindings :one
SELECT COUNT(*) FROM cve_findings;

-- name: CountCVEFindingsFiltered :one
SELECT COUNT(*) FROM cve_findings
WHERE ($1 = '' OR severity::text = $1)
  AND ($2 = '' OR image_name ILIKE '%' || $2 || '%' OR cve_id ILIKE '%' || $2 || '%');

-- name: ListCVEFindingsBySeverity :many
SELECT * FROM cve_findings WHERE severity = $1 ORDER BY scan_date DESC LIMIT $2 OFFSET $3;

-- name: GetCVESummaryFiltered :one
SELECT
    COUNT(*) FILTER (WHERE severity = 'critical') AS critical_count,
    COUNT(*) FILTER (WHERE severity = 'high') AS high_count,
    COUNT(*) FILTER (WHERE severity = 'medium') AS medium_count,
    COUNT(*) FILTER (WHERE severity = 'low') AS low_count
FROM cve_findings
WHERE ($1 = '' OR severity::text = $1)
  AND ($2 = '' OR image_name ILIKE '%' || $2 || '%' OR cve_id ILIKE '%' || $2 || '%');

-- name: UpsertCVEFinding :one
INSERT INTO cve_findings (image_name, image_tag, cve_id, severity, package_name, installed_version, fixed_version, source, scan_date, raw_payload)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
ON CONFLICT (image_name, image_tag, cve_id, source) DO UPDATE SET
    severity = EXCLUDED.severity,
    package_name = EXCLUDED.package_name,
    installed_version = EXCLUDED.installed_version,
    fixed_version = EXCLUDED.fixed_version,
    scan_date = EXCLUDED.scan_date,
    raw_payload = EXCLUDED.raw_payload
RETURNING *;

-- name: GetCVESummary :one
SELECT
    COUNT(*) FILTER (WHERE severity = 'critical') AS critical_count,
    COUNT(*) FILTER (WHERE severity = 'high') AS high_count,
    COUNT(*) FILTER (WHERE severity = 'medium') AS medium_count,
    COUNT(*) FILTER (WHERE severity = 'low') AS low_count
FROM cve_findings;

-- name: DeleteOldCVEFindings :exec
DELETE FROM cve_findings WHERE created_at < $1;
