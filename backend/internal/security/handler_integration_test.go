package security

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/switchboard/switchboard/internal/db"
	"github.com/switchboard/switchboard/internal/testutil"
)

func TestListCVEsIntegration(t *testing.T) {
	queries, pool := testutil.QueriesAndPool(t)
	ctx := context.Background()
	suffix := uuid.New().String()[:8]
	image := fmt.Sprintf("test/%s-app", suffix)

	_, err := queries.UpsertCVEFinding(ctx, db.UpsertCVEFindingParams{
		ImageName: image,
		ImageTag:  "v1",
		CveID:     "CVE-2024-" + suffix,
		Severity:  "high",
		Source:    "webhook",
		ScanDate:  time.Now(),
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM cve_findings WHERE image_name = $1`, image)
	})

	h := NewHandler(queries)
	req := httptest.NewRequest(http.MethodGet, "/?severity=high&search="+suffix+"&limit=10", nil)
	rec := httptest.NewRecorder()
	h.ListCVEs(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
	out := testutil.DecodeJSON[map[string]interface{}](t, rec.Result())
	items, ok := out["items"].([]interface{})
	if !ok || len(items) == 0 {
		t.Fatalf("expected items, got %#v", out["items"])
	}
	total, ok := out["total"].(float64)
	if !ok || total < 1 {
		t.Fatalf("expected total >= 1, got %v", out["total"])
	}
}

func TestListReportsAndGetReportIntegration(t *testing.T) {
	queries, pool := testutil.QueriesAndPool(t)
	ctx := context.Background()
	suffix := uuid.New().String()[:8]
	appName := "test-app-" + suffix

	report, err := queries.CreateDeploymentReport(ctx, db.CreateDeploymentReportParams{
		AppName:     appName,
		ImageName:   "registry.io/" + appName,
		ImageTag:    "latest",
		Status:      "success",
		RawPayload:  []byte(`{"test":true}`),
		PayloadHash: pgtype.Text{String: suffix, Valid: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM deployment_reports WHERE id = $1`, report.ID)
	})

	h := NewHandler(queries)

	listReq := httptest.NewRequest(http.MethodGet, "/?search="+suffix, nil)
	listRec := httptest.NewRecorder()
	h.ListReports(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("list status %d", listRec.Code)
	}
	listOut := testutil.DecodeJSON[map[string]interface{}](t, listRec.Result())
	if listOut["total"].(float64) < 1 {
		t.Fatal("expected at least one report")
	}

	getReq := httptest.NewRequest(http.MethodGet, "/", nil)
	getRec := httptest.NewRecorder()
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", report.ID.String())
	getReq = getReq.WithContext(context.WithValue(getReq.Context(), chi.RouteCtxKey, rctx))
	h.GetReport(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("get status %d body %s", getRec.Code, getRec.Body.String())
	}
	got := testutil.DecodeJSON[db.DeploymentReport](t, getRec.Result())
	if got.AppName != appName {
		t.Fatalf("app_name %q", got.AppName)
	}
}

func TestOverviewAndExportIntegration(t *testing.T) {
	queries, pool := testutil.QueriesAndPool(t)
	ctx := context.Background()
	suffix := uuid.New().String()[:8]
	image := fmt.Sprintf("test/%s-overview", suffix)

	_, err := queries.UpsertCVEFinding(ctx, db.UpsertCVEFindingParams{
		ImageName:    image,
		ImageTag:     "v1",
		CveID:        "CVE-2024-CRIT-" + suffix,
		Severity:     "critical",
		FixedVersion: pgtype.Text{String: "1.2.3", Valid: true},
		Source:       "webhook",
		ScanDate:     time.Now(),
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM cve_findings WHERE image_name = $1`, image)
	})

	h := NewHandler(queries)

	ovReq := httptest.NewRequest(http.MethodGet, "/?top=5", nil)
	ovRec := httptest.NewRecorder()
	h.Overview(ovRec, ovReq)
	if ovRec.Code != http.StatusOK {
		t.Fatalf("overview status %d body %s", ovRec.Code, ovRec.Body.String())
	}
	ov := testutil.DecodeJSON[map[string]interface{}](t, ovRec.Result())
	stats, ok := ov["stats"].(map[string]interface{})
	if !ok {
		t.Fatalf("missing stats: %#v", ov)
	}
	if stats["critical_count"].(float64) < 1 {
		t.Fatalf("expected critical_count >= 1, got %v", stats["critical_count"])
	}

	exReq := httptest.NewRequest(http.MethodGet, "/?search="+suffix, nil)
	exRec := httptest.NewRecorder()
	h.ExportCVEs(exRec, exReq)
	if exRec.Code != http.StatusOK {
		t.Fatalf("export status %d", exRec.Code)
	}
	ct := exRec.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/csv") {
		t.Fatalf("content-type %q", ct)
	}
	body := exRec.Body.String()
	if !strings.Contains(body, "cve_id") || !strings.Contains(body, "CVE-2024-CRIT-"+suffix) {
		t.Fatalf("unexpected csv body: %s", body)
	}
}

func TestGetReportNotFound(t *testing.T) {
	h := NewHandler(testutil.Queries(t))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", uuid.New().String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	h.GetReport(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status %d, want 404", rec.Code)
	}
}
