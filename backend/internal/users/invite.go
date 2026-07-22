package users

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/switchboard/switchboard/internal/audit"
	"github.com/switchboard/switchboard/internal/auth"
	"github.com/switchboard/switchboard/internal/db"
	"github.com/switchboard/switchboard/internal/settings"
	"golang.org/x/crypto/bcrypt"
)

func (h *Handler) GetUserRoles(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}
	roles, err := h.queries.GetUserRoles(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"server error"}`, http.StatusInternalServerError)
		return
	}
	auth.WriteJSON(w, http.StatusOK, roles)
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email       string   `json:"email"`
		DisplayName string   `json:"display_name"`
		Password    string   `json:"password"`
		RoleIDs     []string `json:"role_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}
	email := auth.NormalizeEmail(body.Email)
	if email == "" || body.Password == "" {
		http.Error(w, `{"error":"email and password are required"}`, http.StatusBadRequest)
		return
	}
	if !strings.Contains(email, "@") {
		http.Error(w, `{"error":"invalid email"}`, http.StatusBadRequest)
		return
	}
	if len(body.Password) < 8 {
		http.Error(w, `{"error":"password must be at least 8 characters"}`, http.StatusBadRequest)
		return
	}
	if len(body.RoleIDs) == 0 {
		http.Error(w, `{"error":"at least one role is required"}`, http.StatusBadRequest)
		return
	}
	if _, err := h.queries.GetUserByEmail(r.Context(), email); err == nil {
		http.Error(w, `{"error":"email already exists"}`, http.StatusConflict)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, `{"error":"create failed"}`, http.StatusInternalServerError)
		return
	}
	displayName := strings.TrimSpace(body.DisplayName)
	user, err := h.queries.CreateUser(r.Context(), db.CreateUserParams{
		Username:     auth.EmailAsUsername(email),
		Email:        email,
		DisplayName:  pgtype.Text{String: displayName, Valid: displayName != ""},
		AuthType:     "local",
		PasswordHash: pgtype.Text{String: string(hash), Valid: true},
	})
	if err != nil {
		http.Error(w, `{"error":"create failed"}`, http.StatusInternalServerError)
		return
	}
	if err := h.assignRoles(r.Context(), user.ID, body.RoleIDs); err != nil {
		http.Error(w, `{"error":"role assignment failed"}`, http.StatusInternalServerError)
		return
	}
	roles, _ := h.queries.GetUserRoles(r.Context(), user.ID)
	h.audit.Log(audit.WithClientIP(r.Context(), r.RemoteAddr), "user.create", "user", user.ID.String(), map[string]interface{}{"email": email})
	auth.WriteJSON(w, http.StatusCreated, toUserDTO(user, roles))
}

func (h *Handler) InviteUser(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email       string   `json:"email"`
		DisplayName string   `json:"display_name"`
		RoleIDs     []string `json:"role_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}
	email := auth.NormalizeEmail(body.Email)
	if email == "" {
		http.Error(w, `{"error":"email is required"}`, http.StatusBadRequest)
		return
	}
	if !strings.Contains(email, "@") {
		http.Error(w, `{"error":"invalid email"}`, http.StatusBadRequest)
		return
	}
	if len(body.RoleIDs) == 0 {
		http.Error(w, `{"error":"at least one role is required"}`, http.StatusBadRequest)
		return
	}
	roleUUIDs, err := parseRoleUUIDs(body.RoleIDs)
	if err != nil {
		http.Error(w, `{"error":"invalid role id"}`, http.StatusBadRequest)
		return
	}
	if _, err := h.queries.GetUserByEmail(r.Context(), email); err == nil {
		http.Error(w, `{"error":"email already exists"}`, http.StatusConflict)
		return
	}
	if _, err := h.queries.GetPendingInvitationByEmail(r.Context(), email); err == nil {
		http.Error(w, `{"error":"a pending invitation already exists for this email"}`, http.StatusConflict)
		return
	}

	token, err := newInviteToken()
	if err != nil {
		http.Error(w, `{"error":"invite failed"}`, http.StatusInternalServerError)
		return
	}
	inviterID, _ := auth.UserIDFromContext(r.Context())
	displayName := strings.TrimSpace(body.DisplayName)
	invitation, err := h.queries.CreateInvitation(r.Context(), db.CreateInvitationParams{
		Email:       email,
		Username:    auth.EmailAsUsername(email),
		DisplayName: pgtype.Text{String: displayName, Valid: displayName != ""},
		RoleIds:     roleUUIDs,
		TokenHash:   hashInviteToken(token),
		InvitedBy:   pgtype.UUID{Bytes: inviterID, Valid: inviterID != uuid.Nil},
		ExpiresAt:   invitationExpiry(),
	})
	if err != nil {
		http.Error(w, `{"error":"invite failed"}`, http.StatusInternalServerError)
		return
	}

	link := inviteURL(h.cfg, token)
	var triggeredBy *uuid.UUID
	if inviterID != uuid.Nil {
		triggeredBy = &inviterID
	}
	emailErr := sendInviteEmail(r.Context(), h.queries, h.cfg, email, link, triggeredBy)
	smtpConfigured := settings.ResolveSMTP(r.Context(), h.queries, h.cfg).Configured()
	roleIDStrings := make([]string, len(roleUUIDs))
	for i, id := range roleUUIDs {
		roleIDStrings[i] = id.String()
	}

	auth.WriteJSON(w, http.StatusCreated, map[string]interface{}{
		"invitation": InvitationDTO{
			ID:        invitation.ID,
			Email:     invitation.Email,
			RoleIDs:   roleIDStrings,
			ExpiresAt: invitation.ExpiresAt,
			CreatedAt: invitation.CreatedAt,
		},
		"invite_url":  link,
		"email_sent":  emailErr == nil && smtpConfigured,
		"email_error": errString(emailErr),
	})
}

func (h *Handler) ListInvitations(w http.ResponseWriter, r *http.Request) {
	invitations, err := h.queries.ListPendingInvitations(r.Context())
	if err != nil {
		http.Error(w, `{"error":"server error"}`, http.StatusInternalServerError)
		return
	}
	out := make([]InvitationDTO, len(invitations))
	for i, inv := range invitations {
		roleIDs := make([]string, len(inv.RoleIds))
		for j, id := range inv.RoleIds {
			roleIDs[j] = id.String()
		}
		out[i] = InvitationDTO{
			ID:        inv.ID,
			Email:     inv.Email,
			RoleIDs:   roleIDs,
			ExpiresAt: inv.ExpiresAt,
			CreatedAt: inv.CreatedAt,
		}
	}
	auth.WriteJSON(w, http.StatusOK, out)
}

func (h *Handler) PreviewInvite(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimSpace(r.URL.Query().Get("token"))
	if token == "" {
		http.Error(w, `{"error":"token required"}`, http.StatusBadRequest)
		return
	}
	invitation, err := h.queries.GetInvitationByTokenHash(r.Context(), hashInviteToken(token))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, `{"error":"invalid or expired invitation"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error":"server error"}`, http.StatusInternalServerError)
		return
	}
	auth.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"email":      invitation.Email,
		"expires_at": invitation.ExpiresAt,
	})
}

func (h *Handler) AcceptInvite(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}
	if body.Token == "" || body.Password == "" {
		http.Error(w, `{"error":"token and password are required"}`, http.StatusBadRequest)
		return
	}
	if len(body.Password) < 8 {
		http.Error(w, `{"error":"password must be at least 8 characters"}`, http.StatusBadRequest)
		return
	}

	invitation, err := h.queries.GetInvitationByTokenHash(r.Context(), hashInviteToken(body.Token))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, `{"error":"invalid or expired invitation"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error":"server error"}`, http.StatusInternalServerError)
		return
	}
	if _, err := h.queries.GetUserByEmail(r.Context(), invitation.Email); err == nil {
		http.Error(w, `{"error":"email already taken"}`, http.StatusConflict)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, `{"error":"accept failed"}`, http.StatusInternalServerError)
		return
	}
	user, err := h.queries.CreateUser(r.Context(), db.CreateUserParams{
		Username:     invitation.Username,
		Email:        invitation.Email,
		DisplayName:  invitation.DisplayName,
		AuthType:     "local",
		PasswordHash: pgtype.Text{String: string(hash), Valid: true},
	})
	if err != nil {
		http.Error(w, `{"error":"accept failed"}`, http.StatusInternalServerError)
		return
	}
	roleIDs := make([]string, len(invitation.RoleIds))
	for i, id := range invitation.RoleIds {
		roleIDs[i] = id.String()
		if err := h.queries.AddUserRole(r.Context(), db.AddUserRoleParams{UserID: user.ID, RoleID: id}); err != nil {
			http.Error(w, `{"error":"role assignment failed"}`, http.StatusInternalServerError)
			return
		}
	}
	_ = h.queries.MarkInvitationAccepted(r.Context(), invitation.ID)
	h.local.FinishLogin(w, r, user.ID, user.Username, true)
}

func newInviteToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
