package harbor

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/switchboard/switchboard/internal/db"
)

func loadFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	return data
}

func TestParseScanningCompleted(t *testing.T) {
	reports, err := ParseDeploymentReports(loadFixture(t, "scanning_completed.json"), "https://harbor.example.com")
	if err != nil {
		t.Fatal(err)
	}
	if len(reports) != 1 {
		t.Fatalf("expected 1 report, got %d", len(reports))
	}
	r := reports[0]
	if r.AppName != "library/nginx" {
		t.Errorf("app_name: got %q", r.AppName)
	}
	if r.ImageTag != "v1.2.0" {
		t.Errorf("image_tag: got %q", r.ImageTag)
	}
	if r.Status != db.DeployStatusPartial {
		t.Errorf("status: got %q want partial (has critical/high)", r.Status)
	}
	if r.CriticalCount != 1 || r.HighCount != 2 || r.MediumCount != 5 || r.LowCount != 10 {
		t.Errorf("counts: crit=%d high=%d med=%d low=%d", r.CriticalCount, r.HighCount, r.MediumCount, r.LowCount)
	}
	if r.TriggeredBy != "admin" {
		t.Errorf("triggered_by: got %q", r.TriggeredBy)
	}
}

func TestParsePushArtifact(t *testing.T) {
	reports, err := ParseDeploymentReports(loadFixture(t, "push_artifact.json"), "")
	if err != nil {
		t.Fatal(err)
	}
	if len(reports) != 1 {
		t.Fatalf("expected 1 report, got %d", len(reports))
	}
	r := reports[0]
	if r.Status != db.DeployStatusSuccess {
		t.Errorf("status: got %q", r.Status)
	}
	if r.AppName != "library/myapp" {
		t.Errorf("app_name: got %q", r.AppName)
	}
	if r.TriggeredBy != "ci-bot" {
		t.Errorf("triggered_by: got %q", r.TriggeredBy)
	}
}

func TestParseScanningFailed(t *testing.T) {
	reports, err := ParseDeploymentReports(loadFixture(t, "scanning_failed.json"), "")
	if err != nil {
		t.Fatal(err)
	}
	if reports[0].Status != db.DeployStatusFailed {
		t.Errorf("status: got %q", reports[0].Status)
	}
}

func TestParseLegacyFlat(t *testing.T) {
	reports, err := ParseDeploymentReports(loadFixture(t, "legacy_flat.json"), "")
	if err != nil {
		t.Fatal(err)
	}
	r := reports[0]
	if r.AppName != "library/legacy-app" {
		t.Errorf("app_name: got %q", r.AppName)
	}
	if r.HighCount != 1 {
		t.Errorf("high_count: got %d", r.HighCount)
	}
	if r.Status != db.DeployStatusPartial {
		t.Errorf("status: got %q want partial", r.Status)
	}
}

func TestNormalizeDeployStatus(t *testing.T) {
	tests := []struct {
		scan     string
		hasHigh  bool
		expected db.DeployStatus
	}{
		{"Success", false, db.DeployStatusSuccess},
		{"Success", true, db.DeployStatusPartial},
		{"Error", false, db.DeployStatusFailed},
		{"failed", false, db.DeployStatusFailed},
		{"running", false, db.DeployStatusPartial},
	}
	for _, tc := range tests {
		got := NormalizeDeployStatus(tc.scan, tc.hasHigh)
		if got != tc.expected {
			t.Errorf("NormalizeDeployStatus(%q, %v) = %q, want %q", tc.scan, tc.hasHigh, got, tc.expected)
		}
	}
}

func TestParseResourceURL(t *testing.T) {
	tests := []struct {
		url, wantName, wantTag string
	}{
		{"harbor.example.com/library/nginx:v1.0", "harbor.example.com/library/nginx", "v1.0"},
		{"harbor.example.com/library/nginx@sha256:abc", "harbor.example.com/library/nginx", "@sha256:abc"},
		{"", "", ""},
	}
	for _, tc := range tests {
		name, tag := ParseResourceURL(tc.url)
		if name != tc.wantName || tag != tc.wantTag {
			t.Errorf("ParseResourceURL(%q) = (%q, %q), want (%q, %q)", tc.url, name, tag, tc.wantName, tc.wantTag)
		}
	}
}
