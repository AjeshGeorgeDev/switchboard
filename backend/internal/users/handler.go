package users

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/switchboard/switchboard/internal/audit"
	"github.com/switchboard/switchboard/internal/auth"
	"github.com/switchboard/switchboard/internal/config"
	"github.com/switchboard/switchboard/internal/db"
	"github.com/switchboard/switchboard/internal/rbac"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	queries  *db.Queries
	enforcer *rbac.Enforcer
	sessions *auth.SessionService
	cfg      config.Config
	local    *auth.LocalHandler
	audit    *audit.Logger
}

func NewHandler(queries *db.Queries, enforcer *rbac.Enforcer, sessions *auth.SessionService, cfg config.Config, local *auth.LocalHandler, auditLog *audit.Logger) *Handler {
	return &Handler{queries: queries, enforcer: enforcer, sessions: sessions, cfg: cfg, local: local, audit: auditLog}
}

func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.queries.ListUsers(r.Context())
	if err != nil {
		http.Error(w, `{"error":"server error"}`, http.StatusInternalServerError)
		return
	}
	out := make([]UserDTO, len(users))
	for i, user := range users {
		roles, err := h.queries.GetUserRoles(r.Context(), user.ID)
		if err != nil {
			http.Error(w, `{"error":"server error"}`, http.StatusInternalServerError)
			return
		}
		out[i] = toUserDTO(user, roles)
	}
	auth.WriteJSON(w, http.StatusOK, out)
}

func (h *Handler) SetUserRoles(w http.ResponseWriter, r *http.Request) {
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
	user, err := h.queries.GetUserByID(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	_ = h.queries.SetUserRoles(r.Context(), id)
	roleNames := make([]string, 0, len(body.RoleIDs))
	for _, rid := range body.RoleIDs {
		roleUUID, err := uuid.Parse(rid)
		if err != nil {
			continue
		}
		_ = h.queries.AddUserRole(r.Context(), db.AddUserRoleParams{UserID: id, RoleID: roleUUID})
		if role, err := h.queries.GetRoleByID(r.Context(), roleUUID); err == nil {
			roleNames = append(roleNames, role.Name)
		}
	}
	_ = rbac.SyncUserRoles(h.enforcer, user.Username, roleNames)
	h.audit.Log(audit.WithClientIP(r.Context(), r.RemoteAddr), "user.roles_update", "user", id.String(), map[string]interface{}{"role_ids": body.RoleIDs})
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ForceLogout(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}
	_ = h.sessions.RevokeAll(r.Context(), id)
	h.audit.Log(audit.WithClientIP(r.Context(), r.RemoteAddr), "user.force_logout", "user", id.String(), nil)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListSessions(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}
	sessions, err := h.queries.ListUserLoginHistory(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"server error"}`, http.StatusInternalServerError)
		return
	}
	auth.WriteJSON(w, http.StatusOK, toSessionHistory(sessions))
}

func (h *Handler) ListProfileLoginHistory(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}
	sessions, err := h.queries.ListUserLoginHistory(r.Context(), userID)
	if err != nil {
		http.Error(w, `{"error":"server error"}`, http.StatusInternalServerError)
		return
	}
	auth.WriteJSON(w, http.StatusOK, toSessionHistory(sessions))
}

func (h *Handler) RevokeSession(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}
	_ = h.queries.RevokeRefreshToken(r.Context(), id)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}
	var body struct {
		Email       *string `json:"email"`
		DisplayName *string `json:"display_name"`
		IsActive    *bool   `json:"is_active"`
		Password    *string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}
	params := db.UpdateUserParams{ID: id}
	if body.Email != nil {
		params.Email = pgtype.Text{String: *body.Email, Valid: true}
	}
	if body.DisplayName != nil {
		params.DisplayName = pgtype.Text{String: *body.DisplayName, Valid: true}
	}
	if body.IsActive != nil {
		params.IsActive = pgtype.Bool{Bool: *body.IsActive, Valid: true}
	}
	if body.Password != nil {
		hash, err := bcrypt.GenerateFromPassword([]byte(*body.Password), bcrypt.DefaultCost)
		if err == nil {
			params.PasswordHash = pgtype.Text{String: string(hash), Valid: true}
			params.AuthType = db.NullAuthType{AuthType: "local", Valid: true}
		}
	}
	user, err := h.queries.UpdateUser(r.Context(), params)
	if err != nil {
		http.Error(w, `{"error":"update failed"}`, http.StatusInternalServerError)
		return
	}
	h.audit.Log(audit.WithClientIP(r.Context(), r.RemoteAddr), "user.update", "user", id.String(), map[string]interface{}{
		"email": body.Email, "display_name": body.DisplayName, "is_active": body.IsActive, "password_changed": body.Password != nil,
	})
	auth.WriteJSON(w, http.StatusOK, user)
}

// Roles

func (h *Handler) ListRoles(w http.ResponseWriter, r *http.Request) {
	roles, err := h.queries.ListRoles(r.Context())
	if err != nil {
		http.Error(w, `{"error":"server error"}`, http.StatusInternalServerError)
		return
	}
	auth.WriteJSON(w, http.StatusOK, roles)
}

func (h *Handler) CreateRole(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}
	role, err := h.queries.CreateRole(r.Context(), db.CreateRoleParams{Name: body.Name, Description: pgtype.Text{String: body.Description, Valid: body.Description != ""}})
	if err != nil {
		http.Error(w, `{"error":"create failed"}`, http.StatusInternalServerError)
		return
	}
	h.audit.Log(audit.WithClientIP(r.Context(), r.RemoteAddr), "role.create", "role", role.ID.String(), map[string]interface{}{"name": body.Name})
	auth.WriteJSON(w, http.StatusCreated, role)
}

func (h *Handler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}
	var body struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}
	role, err := h.queries.UpdateRole(r.Context(), db.UpdateRoleParams{ID: id, Name: body.Name, Description: pgtype.Text{String: body.Description, Valid: true}})
	if err != nil {
		http.Error(w, `{"error":"update failed"}`, http.StatusInternalServerError)
		return
	}
	h.audit.Log(audit.WithClientIP(r.Context(), r.RemoteAddr), "role.update", "role", id.String(), map[string]interface{}{"name": body.Name})
	auth.WriteJSON(w, http.StatusOK, role)
}

func (h *Handler) DeleteRole(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}
	_ = h.queries.DeleteRole(r.Context(), id)
	h.audit.Log(audit.WithClientIP(r.Context(), r.RemoteAddr), "role.delete", "role", id.String(), nil)
	w.WriteHeader(http.StatusNoContent)
}

// OIDC providers

func (h *Handler) ListOIDCProviders(w http.ResponseWriter, r *http.Request) {
	providers, err := h.queries.ListOIDCProviders(r.Context())
	if err != nil {
		http.Error(w, `{"error":"server error"}`, http.StatusInternalServerError)
		return
	}
	auth.WriteJSON(w, http.StatusOK, providers)
}

func (h *Handler) CreateOIDCProvider(w http.ResponseWriter, r *http.Request) {
	var body db.CreateOIDCProviderParams
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}
	p, err := h.queries.CreateOIDCProvider(r.Context(), body)
	if err != nil {
		http.Error(w, `{"error":"create failed"}`, http.StatusInternalServerError)
		return
	}
	h.audit.Log(audit.WithClientIP(r.Context(), r.RemoteAddr), "oidc_provider.create", "oidc_provider", p.ID.String(), map[string]interface{}{"name": body.Name})
	auth.WriteJSON(w, http.StatusCreated, p)
}

func (h *Handler) UpdateOIDCProvider(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}
	var body struct {
		DisplayName    string   `json:"display_name"`
		IssuerUrl      string   `json:"issuer_url"`
		ClientID       string   `json:"client_id"`
		ClientSecret   string   `json:"client_secret"`
		Scopes         []string `json:"scopes"`
		AutoProvision  bool     `json:"auto_provision"`
		DefaultRoleID  *string  `json:"default_role_id"`
		IsActive       bool     `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}
	var defaultRole pgtype.UUID
	if body.DefaultRoleID != nil {
		if rid, err := uuid.Parse(*body.DefaultRoleID); err == nil {
			defaultRole = pgtype.UUID{Bytes: rid, Valid: true}
		}
	}
	p, err := h.queries.UpdateOIDCProvider(r.Context(), db.UpdateOIDCProviderParams{
		ID:            id,
		DisplayName:   body.DisplayName,
		IssuerUrl:     body.IssuerUrl,
		ClientID:      body.ClientID,
		ClientSecret:  body.ClientSecret,
		Scopes:        body.Scopes,
		AutoProvision: body.AutoProvision,
		DefaultRoleID: defaultRole,
		IsActive:      body.IsActive,
	})
	if err != nil {
		http.Error(w, `{"error":"update failed"}`, http.StatusInternalServerError)
		return
	}
	h.audit.Log(audit.WithClientIP(r.Context(), r.RemoteAddr), "oidc_provider.update", "oidc_provider", id.String(), map[string]interface{}{"display_name": body.DisplayName, "is_active": body.IsActive})
	auth.WriteJSON(w, http.StatusOK, p)
}

func (h *Handler) DeleteOIDCProvider(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}
	_ = h.queries.DeleteOIDCProvider(r.Context(), id)
	h.audit.Log(audit.WithClientIP(r.Context(), r.RemoteAddr), "oidc_provider.delete", "oidc_provider", id.String(), nil)
	w.WriteHeader(http.StatusNoContent)
}
