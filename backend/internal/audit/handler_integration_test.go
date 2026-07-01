package audit

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/switchboard/switchboard/internal/db"
	"github.com/switchboard/switchboard/internal/testutil"
)

func TestListAuditLogsIntegration(t *testing.T) {
	queries, pool := testutil.QueriesAndPool(t)
	ctx := context.Background()
	logger := New(queries)

	_, err := queries.CreateAuditLog(ctx, db.CreateAuditLogParams{
		ActorUsername: pgtype.Text{String: "tester", Valid: true},
		Action:        "user.update",
		ResourceType:  pgtype.Text{String: "user", Valid: true},
		ResourceID:    pgtype.Text{String: "integration-test", Valid: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM audit_logs WHERE actor_username = 'tester' AND action = 'user.update'`)
	})
	_ = logger

	h := NewHandler(queries)
	req := httptest.NewRequest(http.MethodGet, "/?action=user.update&resource_type=user&limit=10", nil)
	rec := httptest.NewRecorder()
	h.List(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
	out := testutil.DecodeJSON[map[string]interface{}](t, rec.Result())
	if out["total"].(float64) < 1 {
		t.Fatalf("expected audit entries, got %#v", out)
	}
}
