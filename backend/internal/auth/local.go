package auth

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/switchboard/switchboard/internal/db"
	"golang.org/x/crypto/bcrypt"
)

type LocalHandler struct {
	queries  *db.Queries
	sessions *SessionService
	tokens   *TokenService
	audit    interface {
		LogAuth(ctx context.Context, actorID uuid.UUID, actorUsername, action string, r *http.Request, details map[string]interface{})
	}
}

func NewLocalHandler(queries *db.Queries, sessions *SessionService, tokens *TokenService, auditLog interface {
	LogAuth(ctx context.Context, actorID uuid.UUID, actorUsername, action string, r *http.Request, details map[string]interface{})
}) *LocalHandler {
	return &LocalHandler{queries: queries, sessions: sessions, tokens: tokens, audit: auditLog}
}

type loginRequest struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	RememberMe bool   `json:"remember_me"`
}

func (h *LocalHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}

	login := NormalizeEmail(req.Email)
	user, err := h.queries.GetUserByEmail(r.Context(), login)
	if err != nil {
		user, err = h.queries.GetUserByUsername(r.Context(), login)
	}
	if err != nil {
		http.Error(w, `{"error":"invalid credentials"}`, http.StatusUnauthorized)
		return
	}
	if !user.IsActive || user.AuthType != "local" || !user.PasswordHash.Valid {
		http.Error(w, `{"error":"invalid credentials"}`, http.StatusUnauthorized)
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash.String), []byte(req.Password)); err != nil {
		if h.audit != nil {
			h.audit.LogAuth(r.Context(), user.ID, user.Username, "auth.login_failed", r, map[string]interface{}{"method": "local"})
		}
		http.Error(w, `{"error":"invalid credentials"}`, http.StatusUnauthorized)
		return
	}

	h.FinishLogin(w, r, user.ID, user.Username, req.RememberMe)
}

func (h *LocalHandler) FinishLogin(w http.ResponseWriter, r *http.Request, userID uuid.UUID, username string, remember bool) {
	_ = h.queries.UpdateUserLastLogin(r.Context(), userID)
	access, refresh, exp, refreshTTL, err := h.sessions.IssueSession(r.Context(), userID, r.UserAgent(), r.RemoteAddr, remember)
	if err != nil {
		http.Error(w, `{"error":"session error"}`, http.StatusInternalServerError)
		return
	}
	SetAuthCookies(w, access, refresh, exp, refreshTTL, remember)
	if h.audit != nil {
		h.audit.LogAuth(r.Context(), userID, username, "auth.login", r, map[string]interface{}{"method": "local", "remember_me": remember})
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *LocalHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	refresh := RefreshTokenFromRequest(r)
	if refresh == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}
	access, newRefresh, exp, refreshTTL, _, err := h.sessions.RotateRefresh(r.Context(), refresh, r.UserAgent(), r.RemoteAddr)
	if err != nil {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}
	remember := refreshTTL > h.tokens.SessionRefreshTTL()
	SetAuthCookies(w, access, newRefresh, exp, refreshTTL, remember)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *LocalHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var userID uuid.UUID
	var username string
	if refresh := RefreshTokenFromRequest(r); refresh != "" {
		if row, err := h.queries.GetRefreshTokenByHash(r.Context(), HashToken(refresh)); err == nil {
			userID = row.UserID
			if user, err := h.queries.GetUserByID(r.Context(), userID); err == nil {
				username = user.Username
			}
			_ = h.queries.RevokeRefreshToken(r.Context(), row.ID)
		}
	}
	if h.audit != nil && userID != uuid.Nil {
		h.audit.LogAuth(r.Context(), userID, username, "auth.logout", r, nil)
	}
	ClearAuthCookies(w)
	w.WriteHeader(http.StatusNoContent)
}

func (h *LocalHandler) Me(ctx context.Context, userID uuid.UUID) (map[string]interface{}, error) {
	user, err := h.queries.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	roles, err := h.queries.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, err
	}
	roleNames := make([]string, len(roles))
	for i, role := range roles {
		roleNames[i] = role.Name
	}
	return map[string]interface{}{
		"id":           user.ID,
		"username":     user.Username,
		"email":        user.Email,
		"display_name": user.DisplayName.String,
		"auth_type":    user.AuthType,
		"roles":        roleNames,
	}, nil
}
