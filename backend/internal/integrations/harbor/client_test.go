package harbor

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/switchboard/switchboard/internal/config"
)

func TestEncodeHarborRepository(t *testing.T) {
	if got := encodeHarborRepository("nginx"); got != "nginx" {
		t.Fatalf("got %q", got)
	}
	if got := encodeHarborRepository("group/nginx"); got != "group%2Fnginx" {
		t.Fatalf("got %q", got)
	}
}

func TestSplitRepoFullName(t *testing.T) {
	p, r := SplitRepoFullName("library/nginx")
	if p != "library" || r != "nginx" {
		t.Fatalf("got %q %q", p, r)
	}
	p, r = SplitRepoFullName("library/team/api")
	if p != "library" || r != "team/api" {
		t.Fatalf("got %q %q", p, r)
	}
}

func TestParseVulnerabilityAdditionMIMEMap(t *testing.T) {
	body := []byte(`{
		"application/vnd.security.vulnerability.report; version=1.1": {
			"severity": "Critical",
			"vulnerabilities": [
				{
					"id": "CVE-2024-1",
					"package": "openssl",
					"version": "1.1.1",
					"fix_version": "1.1.2",
					"severity": "Critical"
				}
			]
		}
	}`)
	findings, err := parseVulnerabilityAddition(body)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 || findings[0].CVEID != "CVE-2024-1" || findings[0].Package != "openssl" {
		t.Fatalf("unexpected findings: %+v", findings)
	}
}

func TestParseVulnerabilityAdditionDirectReport(t *testing.T) {
	body := []byte(`{
		"vulnerabilities": [
			{"id": "CVE-2024-2", "package": "curl", "version": "7.0", "severity": "High"}
		]
	}`)
	findings, err := parseVulnerabilityAddition(body)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 || findings[0].CVEID != "CVE-2024-2" {
		t.Fatalf("unexpected findings: %+v", findings)
	}
}

func TestFetchArtifactVulnerabilities(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2.0/projects/library/repositories/nginx/artifacts/sha256:abc/additions/vulnerabilities" {
			t.Fatalf("path: %s", r.URL.Path)
		}
		if r.Header.Get("X-Accept-Vulnerabilities") == "" {
			t.Fatal("missing accept vulnerabilities header")
		}
		if r.Header.Get("Authorization") == "" {
			t.Fatal("missing auth")
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			vulnReportMIME: harborVulnReport{
				Vulnerabilities: []struct {
					ID          string   `json:"id"`
					Package     string   `json:"package"`
					Version     string   `json:"version"`
					FixVersion  string   `json:"fix_version"`
					Severity    string   `json:"severity"`
					Description string   `json:"description"`
					Links       []string `json:"links"`
				}{
					{ID: "CVE-TEST", Package: "busybox", Version: "1.0", Severity: "High"},
				},
			},
		})
	}))
	defer srv.Close()

	c := NewClient(config.Config{
		HarborURL:   srv.URL,
		HarborUser:  "robot$lib",
		HarborToken: "secret",
	})
	findings, err := c.FetchArtifactVulnerabilities(t.Context(), "library", "nginx", "sha256:abc")
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 || findings[0].CVEID != "CVE-TEST" {
		t.Fatalf("findings: %+v", findings)
	}
}

func TestFetchArtifactVulnerabilitiesNotConfigured(t *testing.T) {
	c := NewClient(config.Config{})
	findings, err := c.FetchArtifactVulnerabilities(t.Context(), "p", "r", "sha256:x")
	if err != nil || findings != nil {
		t.Fatalf("expected nil,nil got %v %v", findings, err)
	}
}

func TestSetHarborAuth(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "http://example", nil)
	if err := setHarborAuth(req, config.Config{HarborUser: "robot$lib", HarborToken: "secret"}); err != nil {
		t.Fatal(err)
	}
	if !stringsHasPrefix(req.Header.Get("Authorization"), "Basic ") {
		t.Fatalf("expected basic auth, got %q", req.Header.Get("Authorization"))
	}

	req2, _ := http.NewRequest(http.MethodGet, "http://example", nil)
	if err := setHarborAuth(req2, config.Config{HarborToken: "user:pass"}); err != nil {
		t.Fatal(err)
	}
	if !stringsHasPrefix(req2.Header.Get("Authorization"), "Basic ") {
		t.Fatalf("expected basic auth from combined token, got %q", req2.Header.Get("Authorization"))
	}

	req3, _ := http.NewRequest(http.MethodGet, "http://example", nil)
	if err := setHarborAuth(req3, config.Config{HarborToken: "token-only"}); err == nil {
		t.Fatal("expected error for secret-only token without user")
	}
}

func stringsHasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
