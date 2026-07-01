package jobs

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/hibiken/asynq"
	"github.com/switchboard/switchboard/internal/config"
	"github.com/switchboard/switchboard/internal/db"
	"github.com/switchboard/switchboard/internal/notifications"
	"github.com/switchboard/switchboard/internal/testutil"
)

func loadHarborFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "integrations", "harbor", "testdata", name))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	return data
}

func TestHandleHarborWebhookIntegration(t *testing.T) {
	queries, pool := testutil.QueriesAndPool(t)
	cfg := config.Config{HarborURL: "https://harbor.example.com"}
	notify := notifications.NewService(queries, cfg, nil)
	p := &Processor{queries: queries, cfg: cfg, notify: notify}

	body := loadHarborFixture(t, "push_artifact.json")
	envelope, _ := json.Marshal(taskEnvelope{Body: body})
	task := asynq.NewTask(TypeProcessHarborWebhook, envelope)

	if err := p.handleHarborWebhook(context.Background(), task); err != nil {
		t.Fatal(err)
	}

	reports, err := queries.ListDeploymentReportsFiltered(context.Background(), db.ListDeploymentReportsFilteredParams{
		Column1: "library/myapp",
		Column2: "",
		Limit:   10,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(reports) == 0 {
		t.Fatal("expected deployment report from harbor webhook")
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM deployment_reports WHERE app_name = 'library/myapp' AND triggered_by = 'ci-bot'`)
	})
}

func TestHandleTrivyWebhookIntegration(t *testing.T) {
	queries, pool := testutil.QueriesAndPool(t)
	cfg := config.Config{}
	notify := notifications.NewService(queries, cfg, nil)
	p := &Processor{queries: queries, cfg: cfg, notify: notify}

	payload := []byte(`{
		"artifact_name": "registry.io/test-trivy-app:1.0.0",
		"Results": [{
			"Vulnerabilities": [{
				"VulnerabilityID": "CVE-2024-TEST",
				"Severity": "HIGH",
				"PkgName": "openssl",
				"InstalledVersion": "1.1.1",
				"FixedVersion": "1.1.2"
			}]
		}]
	}`)
	task := asynq.NewTask(TypeProcessTrivyWebhook, payload)
	if err := p.handleTrivyWebhook(context.Background(), task); err != nil {
		t.Fatal(err)
	}

	findings, err := queries.ListCVEFindingsFiltered(context.Background(), db.ListCVEFindingsFilteredParams{
		Column1: "",
		Column2: "CVE-2024-TEST",
		Limit:   10,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) == 0 {
		t.Fatal("expected CVE finding from trivy webhook")
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM cve_findings WHERE cve_id = 'CVE-2024-TEST'`)
	})
}
