package harbor

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/switchboard/switchboard/internal/config"
	"github.com/switchboard/switchboard/internal/settings"
)

const vulnReportMIME = "application/vnd.security.vulnerability.report; version=1.1"

type Client struct {
	cred   settings.HarborConfig
	client *http.Client
}

func NewClient(cred settings.HarborConfig) *Client {
	return &Client{
		cred: cred,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// NewClientFromEnv builds a client from process env only (no DB). Prefer NewClient with ResolveHarbor.
func NewClientFromEnv(cfg config.Config) *Client {
	return NewClient(settings.HarborConfig{
		URL:           cfg.HarborURL,
		User:          cfg.HarborUser,
		Token:         cfg.HarborToken,
		WebhookSecret: cfg.HarborWebhookSecret,
	})
}

func (c *Client) Configured() bool {
	return c.cred.APIConfigured()
}

// Ping verifies Harbor URL + credentials with an authenticated API call.
func (c *Client) Ping(ctx context.Context) error {
	if !c.Configured() {
		return fmt.Errorf("Harbor URL and credentials are required")
	}
	base := strings.TrimRight(strings.TrimSpace(c.cred.URL), "/")
	// projects listing works for typical project robots; users/current is less reliable for robots.
	endpoint := base + "/api/v2.0/projects?page_size=1"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	if err := setHarborAuth(req, c.cred); err != nil {
		return err
	}
	res, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("unreachable: %w", err)
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	if res.StatusCode == http.StatusUnauthorized || res.StatusCode == http.StatusForbidden {
		return fmt.Errorf("%s — check robot username and token", res.Status)
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("harbor API %s: %s", res.Status, truncate(body, 160))
	}
	return nil
}

// Finding is a normalized vulnerability from Harbor's artifact additions API.
type Finding struct {
	CVEID            string
	Severity         string
	Package          string
	InstalledVersion string
	FixedVersion     string
	Description      string
	Raw              json.RawMessage
}

type harborVulnReport struct {
	GeneratedAt     string `json:"generated_at"`
	Severity        string `json:"severity"`
	Vulnerabilities []struct {
		ID          string   `json:"id"`
		Package     string   `json:"package"`
		Version     string   `json:"version"`
		FixVersion  string   `json:"fix_version"`
		Severity    string   `json:"severity"`
		Description string   `json:"description"`
		Links       []string `json:"links"`
	} `json:"vulnerabilities"`
}

// FetchArtifactVulnerabilities loads per-CVE details for an artifact digest.
// project is the Harbor project/namespace; repository is the repo path within the project
// (may contain slashes for nested repos).
func (c *Client) FetchArtifactVulnerabilities(ctx context.Context, project, repository, digest string) ([]Finding, error) {
	if !c.Configured() {
		return nil, nil
	}
	project = strings.TrimSpace(project)
	repository = strings.TrimSpace(repository)
	digest = strings.TrimSpace(digest)
	if project == "" || repository == "" || digest == "" {
		return nil, fmt.Errorf("project, repository, and digest are required")
	}

	base := strings.TrimRight(strings.TrimSpace(c.cred.URL), "/")
	endpoint := fmt.Sprintf(
		"%s/api/v2.0/projects/%s/repositories/%s/artifacts/%s/additions/vulnerabilities",
		base,
		url.PathEscape(project),
		encodeHarborRepository(repository),
		url.PathEscape(digest),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Accept-Vulnerabilities", vulnReportMIME)
	if err := setHarborAuth(req, c.cred); err != nil {
		return nil, err
	}

	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if res.StatusCode == http.StatusUnauthorized || res.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("%s — check Harbor user/token (robot Basic auth)", res.Status)
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("harbor vulnerabilities API %s: %s", res.Status, truncate(body, 200))
	}

	return parseVulnerabilityAddition(body)
}

// encodeHarborRepository encodes a repository path so "/" becomes "%2F" while
// other special characters are path-escaped.
func encodeHarborRepository(repository string) string {
	parts := strings.Split(repository, "/")
	for i, part := range parts {
		parts[i] = url.PathEscape(part)
	}
	return strings.Join(parts, "%2F")
}

func setHarborAuth(req *http.Request, cred settings.HarborConfig) error {
	user := strings.TrimSpace(cred.User)
	token := strings.TrimSpace(cred.Token)
	if token == "" {
		return fmt.Errorf("Harbor token is empty")
	}

	var basic string
	switch {
	case user != "":
		basic = user + ":" + token
	case strings.Contains(token, ":"):
		basic = token
	default:
		return fmt.Errorf("set Harbor user to the robot name and token to the secret only")
	}
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(basic)))
	return nil
}

func parseVulnerabilityAddition(body []byte) ([]Finding, error) {
	var byMIME map[string]json.RawMessage
	if err := json.Unmarshal(body, &byMIME); err == nil && len(byMIME) > 0 {
		if raw, ok := byMIME[vulnReportMIME]; ok {
			return parseVulnReport(raw)
		}
		for _, raw := range byMIME {
			findings, err := parseVulnReport(raw)
			if err == nil && len(findings) > 0 {
				return findings, nil
			}
			if err == nil {
				return findings, nil
			}
		}
	}
	return parseVulnReport(body)
}

func parseVulnReport(raw []byte) ([]Finding, error) {
	var report harborVulnReport
	if err := json.Unmarshal(raw, &report); err != nil {
		return nil, err
	}
	out := make([]Finding, 0, len(report.Vulnerabilities))
	for _, v := range report.Vulnerabilities {
		if strings.TrimSpace(v.ID) == "" {
			continue
		}
		itemRaw, _ := json.Marshal(v)
		out = append(out, Finding{
			CVEID:            v.ID,
			Severity:         v.Severity,
			Package:          v.Package,
			InstalledVersion: v.Version,
			FixedVersion:     v.FixVersion,
			Description:      v.Description,
			Raw:              itemRaw,
		})
	}
	return out, nil
}

func truncate(b []byte, n int) string {
	if len(b) <= n {
		return string(b)
	}
	return string(b[:n]) + "…"
}

// SplitRepoFullName splits "project/repo" or "project/group/repo" into project and repository path.
func SplitRepoFullName(full string) (project, repository string) {
	full = strings.TrimSpace(full)
	if full == "" {
		return "", ""
	}
	parts := strings.SplitN(full, "/", 2)
	if len(parts) == 1 {
		return parts[0], parts[0]
	}
	return parts[0], parts[1]
}

func (c *Client) ListRepositories(ctx context.Context) ([]string, error) {
	if c.cred.URL == "" {
		return []string{}, nil
	}
	_ = ctx
	return []string{}, nil
}
