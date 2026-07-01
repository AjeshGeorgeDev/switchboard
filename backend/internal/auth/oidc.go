package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/switchboard/switchboard/internal/config"
	"github.com/switchboard/switchboard/internal/db"
	"golang.org/x/oauth2"
)

type OIDCHandler struct {
	cfg      config.Config
	queries  *db.Queries
	sessions *SessionService
	tokens   *TokenService
	states   sync.Map
	audit    interface {
		LogAuth(ctx context.Context, actorID uuid.UUID, actorUsername, action string, r *http.Request, details map[string]interface{})
	}
}

func NewOIDCHandler(cfg config.Config, queries *db.Queries, sessions *SessionService, tokens *TokenService, auditLog interface {
	LogAuth(ctx context.Context, actorID uuid.UUID, actorUsername, action string, r *http.Request, details map[string]interface{})
}) *OIDCHandler {
	return &OIDCHandler{cfg: cfg, queries: queries, sessions: sessions, tokens: tokens, audit: auditLog}
}

func (h *OIDCHandler) ListProviders(w http.ResponseWriter, r *http.Request) {
	providers, err := h.queries.ListActiveOIDCProviders(r.Context())
	if err != nil {
		http.Error(w, `{"error":"server error"}`, http.StatusInternalServerError)
		return
	}
	type providerResp struct {
		Name        string `json:"name"`
		DisplayName string `json:"display_name"`
	}
	out := make([]providerResp, len(providers))
	for i, p := range providers {
		out[i] = providerResp{Name: p.Name, DisplayName: p.DisplayName}
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

func (h *OIDCHandler) Login(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "provider")
	provider, oauthCfg, oidcProvider, err := h.providerConfig(r.Context(), name)
	if err != nil {
		http.Error(w, `{"error":"provider not found"}`, http.StatusNotFound)
		return
	}

	state := randomState()
	verifier := oauth2.GenerateVerifier()
	h.states.Store(state, verifier)

	authURL := oauthCfg.AuthCodeURL(state, oidc.Nonce(randomState()), oauth2.S256ChallengeOption(verifier))
	_ = provider
	_ = oidcProvider
	http.Redirect(w, r, authURL, http.StatusFound)
}

func (h *OIDCHandler) Callback(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "provider")
	state := r.URL.Query().Get("state")
	verifierVal, ok := h.states.LoadAndDelete(state)
	if !ok {
		http.Error(w, `{"error":"invalid state"}`, http.StatusBadRequest)
		return
	}
	verifier, _ := verifierVal.(string)

	provider, oauthCfg, oidcProvider, err := h.providerConfig(r.Context(), name)
	if err != nil {
		http.Error(w, `{"error":"provider not found"}`, http.StatusNotFound)
		return
	}

	token, err := oauthCfg.Exchange(r.Context(), r.URL.Query().Get("code"), oauth2.VerifierOption(verifier))
	if err != nil {
		http.Error(w, `{"error":"token exchange failed"}`, http.StatusUnauthorized)
		return
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		http.Error(w, `{"error":"missing id_token"}`, http.StatusUnauthorized)
		return
	}
	idToken, err := oidcProvider.Verifier(&oidc.Config{ClientID: provider.ClientID}).Verify(r.Context(), rawIDToken)
	if err != nil {
		http.Error(w, `{"error":"invalid id_token"}`, http.StatusUnauthorized)
		return
	}

	var claims struct {
		Sub   string `json:"sub"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := idToken.Claims(&claims); err != nil {
		http.Error(w, `{"error":"invalid claims"}`, http.StatusUnauthorized)
		return
	}

	user, err := h.queries.GetUserByOIDC(r.Context(), db.GetUserByOIDCParams{
		OidcProvider: pgtype.Text{String: name, Valid: true},
		OidcSubject:  pgtype.Text{String: claims.Sub, Valid: true},
	})
	if err != nil {
		if !provider.AutoProvision {
			http.Error(w, `{"error":"account not provisioned"}`, http.StatusForbidden)
			return
		}
		username := fmt.Sprintf("%s_%s", name, claims.Sub[:8])
		user, err = h.queries.CreateUser(r.Context(), db.CreateUserParams{
			Username:     username,
			Email:        claims.Email,
			DisplayName:  pgtype.Text{String: claims.Name, Valid: claims.Name != ""},
			AuthType:     "oidc",
			OidcProvider: pgtype.Text{String: name, Valid: true},
			OidcSubject:  pgtype.Text{String: claims.Sub, Valid: true},
			OidcEmail:    pgtype.Text{String: claims.Email, Valid: claims.Email != ""},
		})
		if err != nil {
			http.Error(w, `{"error":"provision failed"}`, http.StatusInternalServerError)
			return
		}
		if provider.DefaultRoleID.Valid {
			_ = h.queries.AddUserRole(r.Context(), db.AddUserRoleParams{
				UserID: user.ID,
				RoleID: uuid.UUID(provider.DefaultRoleID.Bytes),
			})
		}
	}

	_ = h.queries.UpdateUserLastLogin(r.Context(), user.ID)
	access, refresh, exp, refreshTTL, err := h.sessions.IssueSession(r.Context(), user.ID, r.UserAgent(), r.RemoteAddr, true)
	if err != nil {
		http.Error(w, `{"error":"session error"}`, http.StatusInternalServerError)
		return
	}
	SetAuthCookies(w, access, refresh, exp, refreshTTL, true)
	if h.audit != nil {
		h.audit.LogAuth(r.Context(), user.ID, user.Username, "auth.login", r, map[string]interface{}{"method": "oidc", "provider": name})
	}
	http.Redirect(w, r, h.cfg.AppBaseURL+"/", http.StatusFound)
}

func (h *OIDCHandler) providerConfig(ctx context.Context, name string) (db.OidcProvider, *oauth2.Config, *oidc.Provider, error) {
	provider, err := h.queries.GetOIDCProviderByName(ctx, name)
	if err != nil || !provider.IsActive {
		return db.OidcProvider{}, nil, nil, fmt.Errorf("not found")
	}
	oidcProvider, err := oidc.NewProvider(ctx, provider.IssuerUrl)
	if err != nil {
		return db.OidcProvider{}, nil, nil, err
	}
	redirectURL := fmt.Sprintf("%s/api/auth/oidc/%s/callback", h.cfg.AppBaseURL, name)
	oauthCfg := &oauth2.Config{
		ClientID:     provider.ClientID,
		ClientSecret: provider.ClientSecret,
		RedirectURL:  redirectURL,
		Endpoint:     oidcProvider.Endpoint(),
		Scopes:       provider.Scopes,
	}
	return provider, oauthCfg, oidcProvider, nil
}

func randomState() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}
