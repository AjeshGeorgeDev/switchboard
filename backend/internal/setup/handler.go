package setup

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/switchboard/switchboard/internal/auth"
	"github.com/switchboard/switchboard/internal/db"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrSetupComplete = errors.New("setup already complete")
	ErrInvalidInput  = errors.New("invalid input")
)

type Handler struct {
	pool     *pgxpool.Pool
	queries  *db.Queries
	local    *auth.LocalHandler
}

func NewHandler(pool *pgxpool.Pool, queries *db.Queries, local *auth.LocalHandler) *Handler {
	return &Handler{pool: pool, queries: queries, local: local}
}

func IsComplete(ctx context.Context, queries *db.Queries) (bool, error) {
	return queries.HasAdminUser(ctx)
}

func (h *Handler) Status(w http.ResponseWriter, r *http.Request) {
	complete, err := IsComplete(r.Context(), h.queries)
	if err != nil {
		http.Error(w, `{"error":"server error"}`, http.StatusInternalServerError)
		return
	}
	auth.WriteJSON(w, http.StatusOK, map[string]bool{"complete": complete})
}

type setupRequest struct {
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	Password    string `json:"password"`
}

func (h *Handler) Complete(w http.ResponseWriter, r *http.Request) {
	complete, err := IsComplete(r.Context(), h.queries)
	if err != nil {
		http.Error(w, `{"error":"server error"}`, http.StatusInternalServerError)
		return
	}
	if complete {
		http.Error(w, `{"error":"setup already complete"}`, http.StatusConflict)
		return
	}

	var req setupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}
	if err := validateSetupRequest(req); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	userID, err := h.createAdminUser(r.Context(), req)
	if err != nil {
		if errors.Is(err, ErrSetupComplete) {
			http.Error(w, `{"error":"setup already complete"}`, http.StatusConflict)
			return
		}
		http.Error(w, `{"error":"setup failed"}`, http.StatusInternalServerError)
		return
	}

	h.local.FinishLogin(w, r, userID, auth.EmailAsUsername(req.Email), true)
}

func (h *Handler) createAdminUser(ctx context.Context, req setupRequest) (uuid.UUID, error) {
	tx, err := h.pool.Begin(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	defer tx.Rollback(ctx)

	qtx := h.queries.WithTx(tx)

	hasAdmin, err := qtx.HasAdminUser(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	if hasAdmin {
		return uuid.Nil, ErrSetupComplete
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return uuid.Nil, err
	}

	displayName := req.DisplayName
	if displayName == "" {
		displayName = "Administrator"
	}

	user, err := qtx.CreateUser(ctx, db.CreateUserParams{
		Username:     auth.EmailAsUsername(req.Email),
		Email:        auth.NormalizeEmail(req.Email),
		DisplayName:  pgtype.Text{String: displayName, Valid: true},
		AuthType:     "local",
		PasswordHash: pgtype.Text{String: string(hash), Valid: true},
	})
	if err != nil {
		return uuid.Nil, err
	}

	adminRole, err := qtx.GetRoleByName(ctx, "admin")
	if err != nil {
		return uuid.Nil, err
	}

	if err := qtx.AddUserRole(ctx, db.AddUserRoleParams{
		UserID: user.ID,
		RoleID: adminRole.ID,
	}); err != nil {
		return uuid.Nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return uuid.Nil, err
	}

	return user.ID, nil
}

func validateSetupRequest(req setupRequest) error {
	email := auth.NormalizeEmail(req.Email)
	password := req.Password

	if email == "" || password == "" {
		return ErrInvalidInput
	}
	if !strings.Contains(email, "@") {
		return errors.New("invalid email")
	}
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	return nil
}

func BlockIfIncomplete(queries *db.Queries) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			complete, err := IsComplete(r.Context(), queries)
			if err != nil {
				http.Error(w, `{"error":"server error"}`, http.StatusInternalServerError)
				return
			}
			if !complete {
				http.Error(w, `{"error":"setup required"}`, http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
