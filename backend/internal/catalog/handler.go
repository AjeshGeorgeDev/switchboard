package catalog

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/switchboard/switchboard/internal/auth"
	"github.com/switchboard/switchboard/internal/db"
)

type Handler struct {
	queries *db.Queries
}

func NewHandler(queries *db.Queries) *Handler {
	return &Handler{queries: queries}
}

func (h *Handler) ListPublic(w http.ResponseWriter, r *http.Request) {
	apps, err := h.queries.ListPublicApplications(r.Context())
	if err != nil {
		http.Error(w, `{"error":"server error"}`, http.StatusInternalServerError)
		return
	}
	auth.WriteJSON(w, http.StatusOK, apps)
}

func (h *Handler) ListForUser(w http.ResponseWriter, r *http.Request) {
	roles := auth.RolesFromContext(r.Context())
	for _, role := range roles {
		if role == "admin" {
			apps, err := h.queries.ListApplications(r.Context())
			if err != nil {
				http.Error(w, `{"error":"server error"}`, http.StatusInternalServerError)
				return
			}
			auth.WriteJSON(w, http.StatusOK, apps)
			return
		}
	}
	apps, err := h.queries.ListApplicationsForRoles(r.Context(), roles)
	if err != nil {
		http.Error(w, `{"error":"server error"}`, http.StatusInternalServerError)
		return
	}
	auth.WriteJSON(w, http.StatusOK, apps)
}

func (h *Handler) PreviewForRole(w http.ResponseWriter, r *http.Request) {
	role := strings.TrimSpace(r.URL.Query().Get("role"))
	if role == "" {
		http.Error(w, `{"error":"role is required"}`, http.StatusBadRequest)
		return
	}
	apps, err := h.queries.ListApplicationsForRoles(r.Context(), []string{role})
	if err != nil {
		http.Error(w, `{"error":"server error"}`, http.StatusInternalServerError)
		return
	}
	auth.WriteJSON(w, http.StatusOK, apps)
}

func (h *Handler) ListAll(w http.ResponseWriter, r *http.Request) {
	apps, err := h.queries.ListApplications(r.Context())
	if err != nil {
		http.Error(w, `{"error":"server error"}`, http.StatusInternalServerError)
		return
	}
	auth.WriteJSON(w, http.StatusOK, apps)
}

type appRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	IconURL     string   `json:"icon_url"`
	AccessType  string   `json:"access_type"`
	TargetHost  string   `json:"target_host"`
	TargetPort  *int32   `json:"target_port"`
	IsActive    bool     `json:"is_active"`
	IsPublic    bool     `json:"is_public"`
	SortOrder   int32    `json:"sort_order"`
	SectionID   *string  `json:"section_id"`
	RoleIDs     []string `json:"role_ids"`
}

func parseSectionID(raw *string) pgtype.UUID {
	if raw == nil || strings.TrimSpace(*raw) == "" {
		return pgtype.UUID{Valid: false}
	}
	id, err := uuid.Parse(strings.TrimSpace(*raw))
	if err != nil {
		return pgtype.UUID{Valid: false}
	}
	return pgtype.UUID{Bytes: id, Valid: true}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req appRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.TargetHost) == "" {
		http.Error(w, `{"error":"name and target_host are required"}`, http.StatusBadRequest)
		return
	}
	if req.AccessType == "ip_port" && req.TargetPort == nil {
		http.Error(w, `{"error":"target_port is required for ip_port access"}`, http.StatusBadRequest)
		return
	}
	userID, _ := auth.UserIDFromContext(r.Context())
	var port pgtype.Int4
	if req.TargetPort != nil {
		port = pgtype.Int4{Int32: *req.TargetPort, Valid: true}
	}
	app, err := h.queries.CreateApplication(r.Context(), db.CreateApplicationParams{
		Name:        req.Name,
		Description: pgtype.Text{String: req.Description, Valid: req.Description != ""},
		IconUrl:     pgtype.Text{String: req.IconURL, Valid: req.IconURL != ""},
		AccessType:  req.AccessType,
		TargetHost:  req.TargetHost,
		TargetPort:  port,
		IsActive:    req.IsActive,
		IsPublic:    req.IsPublic,
		SortOrder:   req.SortOrder,
		SectionID:   parseSectionID(req.SectionID),
		CreatedBy:   pgtype.UUID{Bytes: userID, Valid: true},
	})
	if err != nil {
		http.Error(w, `{"error":"create failed"}`, http.StatusInternalServerError)
		return
	}
	h.setRoles(r, app.ID, req.RoleIDs)
	auth.WriteJSON(w, http.StatusCreated, app)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}
	var req appRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.TargetHost) == "" {
		http.Error(w, `{"error":"name and target_host are required"}`, http.StatusBadRequest)
		return
	}
	if req.AccessType == "ip_port" && req.TargetPort == nil {
		http.Error(w, `{"error":"target_port is required for ip_port access"}`, http.StatusBadRequest)
		return
	}
	var port pgtype.Int4
	if req.TargetPort != nil {
		port = pgtype.Int4{Int32: *req.TargetPort, Valid: true}
	}
	app, err := h.queries.UpdateApplication(r.Context(), db.UpdateApplicationParams{
		ID:          id,
		Name:        req.Name,
		Description: pgtype.Text{String: req.Description, Valid: true},
		IconUrl:     pgtype.Text{String: req.IconURL, Valid: true},
		AccessType:  req.AccessType,
		TargetHost:  req.TargetHost,
		TargetPort:  port,
		IsActive:    req.IsActive,
		IsPublic:    req.IsPublic,
		SortOrder:   req.SortOrder,
		SectionID:   parseSectionID(req.SectionID),
	})
	if err != nil {
		http.Error(w, `{"error":"update failed"}`, http.StatusInternalServerError)
		return
	}
	if req.RoleIDs != nil {
		h.setRoles(r, id, req.RoleIDs)
	}
	auth.WriteJSON(w, http.StatusOK, app)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}
	_ = h.queries.DeleteApplication(r.Context(), id)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) SetRoles(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}
	var body struct {
		RoleIDs []string `json:"role_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}
	h.setRoles(r, id, body.RoleIDs)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) GetRoles(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}
	roles, err := h.queries.GetApplicationRoles(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"server error"}`, http.StatusInternalServerError)
		return
	}
	auth.WriteJSON(w, http.StatusOK, roles)
}

func (h *Handler) setRoles(r *http.Request, appID uuid.UUID, roleIDs []string) {
	_ = h.queries.SetApplicationRoles(r.Context(), appID)
	for _, rid := range roleIDs {
		roleUUID, err := uuid.Parse(rid)
		if err != nil {
			continue
		}
		_ = h.queries.AddApplicationRole(r.Context(), db.AddApplicationRoleParams{
			ApplicationID: appID,
			RoleID:        roleUUID,
		})
	}
}
