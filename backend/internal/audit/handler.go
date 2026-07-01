package audit

import (
	"net/http"
	"strconv"

	"github.com/switchboard/switchboard/internal/auth"
	"github.com/switchboard/switchboard/internal/db"
)

type Handler struct {
	queries *db.Queries
}

func NewHandler(queries *db.Queries) *Handler {
	return &Handler{queries: queries}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	limit, offset := pagination(r)
	action := r.URL.Query().Get("action")
	resourceType := r.URL.Query().Get("resource_type")
	logs, err := h.queries.ListAuditLogs(r.Context(), db.ListAuditLogsParams{
		Column1: action,
		Column2: resourceType,
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		http.Error(w, `{"error":"server error"}`, http.StatusInternalServerError)
		return
	}
	count, _ := h.queries.CountAuditLogs(r.Context(), db.CountAuditLogsParams{
		Column1: action,
		Column2: resourceType,
	})
	auth.WriteJSON(w, http.StatusOK, map[string]interface{}{"items": logs, "total": count})
}

func pagination(r *http.Request) (int32, int32) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	return int32(limit), int32(offset)
}
