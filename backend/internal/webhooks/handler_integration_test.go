package webhooks

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/switchboard/switchboard/internal/config"
	"github.com/switchboard/switchboard/internal/db"
	"github.com/switchboard/switchboard/internal/testutil"
)

func TestListWebhookEventsIntegration(t *testing.T) {
	queries, pool := testutil.QueriesAndPool(t)
	ctx := context.Background()

	event, err := queries.CreateWebhookEvent(ctx, db.CreateWebhookEventParams{
		Source:  db.WebhookSourceHarbor,
		Payload: []byte(`{"type":"PUSH_ARTIFACT"}`),
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM webhook_events WHERE id = $1`, event.ID)
	})

	h := NewHandler(nil, configWithBase(), queries)
	req := httptest.NewRequest(http.MethodGet, "/?source=harbor&limit=10", nil)
	rec := httptest.NewRecorder()
	h.ListEvents(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
	out := testutil.DecodeJSON[map[string]interface{}](t, rec.Result())
	items := out["items"].([]interface{})
	if len(items) == 0 {
		t.Fatal("expected webhook events")
	}
}

func TestGetWebhookEventIntegration(t *testing.T) {
	queries, pool := testutil.QueriesAndPool(t)
	ctx := context.Background()

	event, err := queries.CreateWebhookEvent(ctx, db.CreateWebhookEventParams{
		Source:  db.WebhookSourceTrivy,
		Payload: []byte(`{"artifact_name":"app:tag"}`),
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM webhook_events WHERE id = $1`, event.ID)
	})

	h := NewHandler(nil, configWithBase(), queries)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", event.ID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	h.GetEvent(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	got := testutil.DecodeJSON[db.WebhookEvent](t, rec.Result())
	if got.Source != db.WebhookSourceTrivy {
		t.Fatalf("source %q", got.Source)
	}
}

func TestGetEventInvalidID(t *testing.T) {
	h := NewHandler(nil, configWithBase(), testutil.Queries(t))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "not-a-uuid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	h.GetEvent(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want 400", rec.Code)
	}
}

func TestGetEventNotFound(t *testing.T) {
	h := NewHandler(nil, configWithBase(), testutil.Queries(t))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", uuid.New().String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	h.GetEvent(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status %d, want 404", rec.Code)
	}
}

func configWithBase() config.Config {
	return config.Config{AppBaseURL: "http://localhost:8080"}
}
