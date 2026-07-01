package security

import (
	"net/http"
	"strconv"

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
	auth.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"items":   findings,
		"total":   count,
		"summary": summary,
	})
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
