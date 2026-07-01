package auth

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/switchboard/switchboard/internal/db"
)

type Middleware struct {
	tokens  *TokenService
	queries *db.Queries
}

func NewMiddleware(tokens *TokenService, queries *db.Queries) *Middleware {
	return &Middleware{tokens: tokens, queries: queries}
}

func (m *Middleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := AccessTokenFromRequest(r)
		if token == "" {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		userID, err := m.tokens.ParseAccessToken(token)
		if err != nil {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		user, err := m.queries.GetUserByID(r.Context(), userID)
		if err != nil || !user.IsActive {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		roles, err := m.queries.GetUserRoles(r.Context(), userID)
		if err != nil {
			http.Error(w, `{"error":"server error"}`, http.StatusInternalServerError)
			return
		}
		roleNames := make([]string, len(roles))
		for i, role := range roles {
			roleNames[i] = role.Name
		}
		ctx := WithUser(r.Context(), userID, user.Username, roleNames)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *Middleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := AccessTokenFromRequest(r)
		if token == "" {
			next.ServeHTTP(w, r)
			return
		}
		userID, err := m.tokens.ParseAccessToken(token)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		user, err := m.queries.GetUserByID(r.Context(), userID)
		if err != nil || !user.IsActive {
			next.ServeHTTP(w, r)
			return
		}
		roles, _ := m.queries.GetUserRoles(r.Context(), userID)
		roleNames := make([]string, len(roles))
		for i, role := range roles {
			roleNames[i] = role.Name
		}
		ctx := WithUser(r.Context(), userID, user.Username, roleNames)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func WriteJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func MeHandler(local *LocalHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := UserIDFromContext(r.Context())
		if !ok {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		me, err := local.Me(r.Context(), userID)
		if err != nil {
			http.Error(w, `{"error":"server error"}`, http.StatusInternalServerError)
			return
		}
		WriteJSON(w, http.StatusOK, me)
	}
}

func GetRolesFromRequest(r *http.Request) []string {
	return RolesFromContext(r.Context())
}

func GetUserID(ctx context.Context) (string, bool) {
	id, ok := UserIDFromContext(ctx)
	if !ok {
		return "", false
	}
	return id.String(), true
}
