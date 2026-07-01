package rbac

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type policy struct {
	Role   string
	Object string
	Action string
}

type Enforcer struct {
	mu       sync.RWMutex
	policies []policy
	pool     *pgxpool.Pool
}

func New(pool *pgxpool.Pool) (*Enforcer, error) {
	e := &Enforcer{pool: pool}
	if err := e.LoadPolicy(context.Background()); err != nil {
		return nil, err
	}
	return e, nil
}

func (e *Enforcer) LoadPolicy(ctx context.Context) error {
	rows, err := e.pool.Query(ctx, `SELECT v0, v1, v2 FROM casbin_rule WHERE ptype = 'p'`)
	if err != nil {
		return fmt.Errorf("load policies: %w", err)
	}
	defer rows.Close()

	var policies []policy
	for rows.Next() {
		var p policy
		if err := rows.Scan(&p.Role, &p.Object, &p.Action); err != nil {
			return err
		}
		policies = append(policies, p)
	}
	e.mu.Lock()
	e.policies = policies
	e.mu.Unlock()
	return rows.Err()
}

func (e *Enforcer) HasPermission(roles []string, obj, act string) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	for _, role := range roles {
		for _, p := range e.policies {
			if p.Role != role {
				continue
			}
			objMatch := p.Object == obj || p.Object == "*"
			actMatch := p.Action == act || p.Action == "*"
			if objMatch && actMatch {
				return true
			}
		}
	}
	return false
}

func RequirePermission(e *Enforcer, getRoles func(*http.Request) []string, obj, act string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			roles := getRoles(r)
			if len(roles) == 0 {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}
			if !e.HasPermission(roles, obj, act) {
				http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func SyncUserRoles(_ *Enforcer, _ string, _ []string) error {
	// Role permissions are role-based via casbin_rule; user-role mapping is in user_roles table.
	return nil
}

func ChiRouteObject(r *http.Request) string {
	route := chi.RouteContext(r.Context()).RoutePattern()
	switch {
	case route == "/api/catalog" || route == "/api/catalog/*":
		return "catalog"
	case route == "/api/security/*" || route == "/api/cve/*" || route == "/api/reports/*":
		return "security"
	case route == "/api/admin/*":
		return "admin"
	case route == "/api/notifications" || route == "/api/notifications/*":
		return "notifications"
	default:
		return "catalog"
	}
}
