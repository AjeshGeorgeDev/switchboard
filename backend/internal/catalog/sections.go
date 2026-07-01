package catalog

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/switchboard/switchboard/internal/auth"
	"github.com/switchboard/switchboard/internal/db"
)

type sectionRequest struct {
	Name      string `json:"name"`
	SortOrder int32  `json:"sort_order"`
}

func (h *Handler) ListSections(w http.ResponseWriter, r *http.Request) {
	sections, err := h.queries.ListCatalogSections(r.Context())
	if err != nil {
		http.Error(w, `{"error":"server error"}`, http.StatusInternalServerError)
		return
	}
	auth.WriteJSON(w, http.StatusOK, sections)
}

func (h *Handler) CreateSection(w http.ResponseWriter, r *http.Request) {
	var req sectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		http.Error(w, `{"error":"name is required"}`, http.StatusBadRequest)
		return
	}
	section, err := h.queries.CreateCatalogSection(r.Context(), db.CreateCatalogSectionParams{
		Name:      strings.TrimSpace(req.Name),
		SortOrder: req.SortOrder,
	})
	if err != nil {
		http.Error(w, `{"error":"create failed"}`, http.StatusInternalServerError)
		return
	}
	auth.WriteJSON(w, http.StatusCreated, section)
}

func (h *Handler) UpdateSection(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}
	var req sectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		http.Error(w, `{"error":"name is required"}`, http.StatusBadRequest)
		return
	}
	section, err := h.queries.UpdateCatalogSection(r.Context(), db.UpdateCatalogSectionParams{
		ID:        id,
		Name:      strings.TrimSpace(req.Name),
		SortOrder: req.SortOrder,
	})
	if err != nil {
		http.Error(w, `{"error":"update failed"}`, http.StatusInternalServerError)
		return
	}
	auth.WriteJSON(w, http.StatusOK, section)
}

func (h *Handler) DeleteSection(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}
	if err := h.queries.DeleteCatalogSection(r.Context(), id); err != nil {
		http.Error(w, `{"error":"delete failed"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
