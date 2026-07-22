package security

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/switchboard/switchboard/internal/auth"
	"github.com/switchboard/switchboard/internal/db"
)

type Handler struct {
	queries *db.Queries
}

func NewHandler(queries *db.Queries) *Handler {
	return &Handler{queries: queries}
}

func (h *Handler) Overview(w http.ResponseWriter, r *http.Request) {
	topN := int32(10)
	if n, err := strconv.Atoi(r.URL.Query().Get("top")); err == nil && n > 0 && n <= 50 {
		topN = int32(n)
	}
	out, err := BuildOverview(r.Context(), h.queries, topN)
	if err != nil {
		http.Error(w, `{"error":"server error"}`, http.StatusInternalServerError)
		return
	}
	auth.WriteJSON(w, http.StatusOK, out)
}

func (h *Handler) ListImages(w http.ResponseWriter, r *http.Request) {
	limit, offset := pagination(r)
	rows, err := h.queries.ListImageRiskRollup(r.Context(), db.ListImageRiskRollupParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		http.Error(w, `{"error":"server error"}`, http.StatusInternalServerError)
		return
	}
	total, _ := h.queries.CountImageRiskRollup(r.Context())
	items := make([]RiskyImage, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapRollupImage(row))
	}
	auth.WriteJSON(w, http.StatusOK, map[string]interface{}{"items": items, "total": total})
}

func (h *Handler) ListCVEs(w http.ResponseWriter, r *http.Request) {
	limit, offset := pagination(r)
	severity := r.URL.Query().Get("severity")
	search := r.URL.Query().Get("search")

	findings, err := h.queries.ListCVEFindingsFiltered(r.Context(), db.ListCVEFindingsFilteredParams{
		Column1: severity,
		Column2: search,
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		http.Error(w, `{"error":"server error"}`, http.StatusInternalServerError)
		return
	}
	count, _ := h.queries.CountCVEFindingsFiltered(r.Context(), db.CountCVEFindingsFilteredParams{
		Column1: severity,
		Column2: search,
	})
	summary, _ := h.queries.GetCVESummaryFiltered(r.Context(), db.GetCVESummaryFilteredParams{
		Column1: severity,
		Column2: search,
	})

	items := make([]map[string]interface{}, 0, len(findings))
	for _, f := range findings {
		fixed := ""
		if f.FixedVersion.Valid {
			fixed = f.FixedVersion.String
		}
		items = append(items, map[string]interface{}{
			"id":                f.ID,
			"image_name":        f.ImageName,
			"image_tag":         f.ImageTag,
			"cve_id":            f.CveID,
			"severity":         f.Severity,
			"package_name":      f.PackageName,
			"installed_version": f.InstalledVersion,
			"fixed_version":     f.FixedVersion,
			"source":            f.Source,
			"scan_date":         f.ScanDate,
			"created_at":        f.CreatedAt,
			"fixable":          isFixable(fixed),
			"age_days":          findingAgeDays(f.CreatedAt),
		})
	}

	auth.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"items":   items,
		"total":   count,
		"summary": summary,
	})
}

func (h *Handler) ExportCVEs(w http.ResponseWriter, r *http.Request) {
	severity := r.URL.Query().Get("severity")
	search := r.URL.Query().Get("search")
	limit := int32(5000)
	if n, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && n > 0 && n <= 20000 {
		limit = int32(n)
	}

	findings, err := h.queries.ListCVEFindingsForExport(r.Context(), db.ListCVEFindingsForExportParams{
		Column1: severity,
		Column2: search,
		Limit:   limit,
	})
	if err != nil {
		http.Error(w, `{"error":"server error"}`, http.StatusInternalServerError)
		return
	}

	filename := fmt.Sprintf("switchboard-cves-%s.csv", time.Now().UTC().Format("20060102"))
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))

	cw := csv.NewWriter(w)
	_ = cw.Write([]string{
		"image_name", "image_tag", "cve_id", "severity", "package_name",
		"installed_version", "fixed_version", "scan_date", "age_days", "fixable",
	})
	for _, f := range findings {
		pkg, installed, fixed := "", "", ""
		if f.PackageName.Valid {
			pkg = f.PackageName.String
		}
		if f.InstalledVersion.Valid {
			installed = f.InstalledVersion.String
		}
		if f.FixedVersion.Valid {
			fixed = f.FixedVersion.String
		}
		_ = cw.Write([]string{
			f.ImageName,
			f.ImageTag,
			f.CveID,
			f.Severity,
			pkg,
			installed,
			fixed,
			f.ScanDate.UTC().Format(time.RFC3339),
			strconv.Itoa(findingAgeDays(f.CreatedAt)),
			strconv.FormatBool(isFixable(fixed)),
		})
	}
	cw.Flush()
}

func (h *Handler) ListReports(w http.ResponseWriter, r *http.Request) {
	limit, offset := pagination(r)
	search := r.URL.Query().Get("search")
	status := r.URL.Query().Get("status")

	reports, err := h.queries.ListDeploymentReportsFiltered(r.Context(), db.ListDeploymentReportsFilteredParams{
		Column1: search,
		Column2: status,
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		http.Error(w, `{"error":"server error"}`, http.StatusInternalServerError)
		return
	}
	count, _ := h.queries.CountDeploymentReportsFiltered(r.Context(), db.CountDeploymentReportsFilteredParams{
		Column1: search,
		Column2: status,
	})
	auth.WriteJSON(w, http.StatusOK, map[string]interface{}{"items": reports, "total": count})
}

func (h *Handler) GetReport(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}
	report, err := h.queries.GetDeploymentReportByID(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	auth.WriteJSON(w, http.StatusOK, report)
}

func pagination(r *http.Request) (int32, int32) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	return int32(limit), int32(offset)
}
