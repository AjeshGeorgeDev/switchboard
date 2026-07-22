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

-- name: GetCVEOverviewStats :one
SELECT
    COUNT(*) FILTER (WHERE severity = 'critical') AS critical_count,
    COUNT(*) FILTER (WHERE severity = 'high') AS high_count,
    COUNT(*) FILTER (WHERE severity = 'medium') AS medium_count,
    COUNT(*) FILTER (WHERE severity = 'low') AS low_count,
    COUNT(*) FILTER (
        WHERE severity = 'critical'
          AND fixed_version IS NOT NULL
          AND BTRIM(fixed_version) <> ''
    ) AS fixable_critical,
    COUNT(*) FILTER (
        WHERE severity IN ('critical', 'high')
          AND fixed_version IS NOT NULL
          AND BTRIM(fixed_version) <> ''
    ) AS fixable_critical_high,
    COUNT(*) FILTER (
        WHERE severity IN ('critical', 'high')
          AND (fixed_version IS NULL OR BTRIM(fixed_version) = '')
    ) AS unfixed_critical_high,
    COUNT(*) FILTER (WHERE created_at >= now() - interval '7 days') AS new_this_week,
    COUNT(*) FILTER (
        WHERE severity IN ('critical', 'high')
          AND created_at >= now() - interval '7 days'
    ) AS aging_lt_7d,
    COUNT(*) FILTER (
        WHERE severity IN ('critical', 'high')
          AND created_at < now() - interval '7 days'
          AND created_at >= now() - interval '30 days'
    ) AS aging_7_to_30d,
    COUNT(*) FILTER (
        WHERE severity IN ('critical', 'high')
          AND created_at < now() - interval '30 days'
    ) AS aging_gt_30d
FROM cve_findings;

-- name: ListTopRiskyImages :many
SELECT
    image_name,
    (ARRAY_AGG(image_tag ORDER BY scan_date DESC))[1]::text AS latest_tag,
    COUNT(*) FILTER (WHERE severity = 'critical')::bigint AS critical_count,
    COUNT(*) FILTER (WHERE severity = 'high')::bigint AS high_count,
    COUNT(*) FILTER (WHERE severity = 'medium')::bigint AS medium_count,
    COUNT(*) FILTER (WHERE severity = 'low')::bigint AS low_count,
    COUNT(*)::bigint AS total_count,
    MIN(created_at) FILTER (WHERE severity = 'critical') AS oldest_critical_at
FROM cve_findings
GROUP BY image_name
ORDER BY critical_count DESC, high_count DESC, total_count DESC
LIMIT $1;

-- name: ListImageRiskRollup :many
SELECT
    image_name,
    (ARRAY_AGG(image_tag ORDER BY scan_date DESC))[1]::text AS latest_tag,
    COUNT(DISTINCT image_tag)::bigint AS tag_count,
    COUNT(*) FILTER (WHERE severity = 'critical')::bigint AS critical_count,
    COUNT(*) FILTER (WHERE severity = 'high')::bigint AS high_count,
    COUNT(*) FILTER (WHERE severity = 'medium')::bigint AS medium_count,
    COUNT(*) FILTER (WHERE severity = 'low')::bigint AS low_count,
    COUNT(*)::bigint AS total_count,
    MIN(created_at) FILTER (WHERE severity = 'critical') AS oldest_critical_at
FROM cve_findings
GROUP BY image_name
ORDER BY critical_count DESC, high_count DESC, total_count DESC
LIMIT $1 OFFSET $2;

-- name: CountImageRiskRollup :one
SELECT COUNT(DISTINCT image_name) FROM cve_findings;

-- name: ListCVEFindingsForExport :many
SELECT * FROM cve_findings
WHERE ($1 = '' OR severity::text = $1)
  AND ($2 = '' OR image_name ILIKE '%' || $2 || '%' OR cve_id ILIKE '%' || $2 || '%')
ORDER BY
    CASE severity
        WHEN 'critical' THEN 1
        WHEN 'high' THEN 2
        WHEN 'medium' THEN 3
        ELSE 4
    END,
    scan_date DESC
LIMIT $3;

-- name: DeleteOldCVEFindings :exec
DELETE FROM cve_findings WHERE created_at < $1;
