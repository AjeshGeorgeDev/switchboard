package trivy

import (
	"context"
	"fmt"

	"github.com/switchboard/switchboard/internal/config"
)

type Finding struct {
	ImageName string
	ImageTag  string
	CVEID     string
	Severity  string
	Package   string
	Installed string
	Fixed     string
}

type Client struct {
	cfg config.Config
}

func NewClient(cfg config.Config) *Client {
	return &Client{cfg: cfg}
}

// FetchAllCVEs pulls findings from a Trivy server when configured.
// Trivy does not expose a global CVE inventory API; scheduled pull requires
// a custom integration. Returns empty results when not configured.
func (c *Client) FetchAllCVEs(ctx context.Context) ([]Finding, error) {
	_ = ctx
	if !c.cfg.CVEPullEnabled {
		return nil, nil
	}
	if c.cfg.TrivyURL == "" || c.cfg.TrivyToken == "" {
		return nil, fmt.Errorf("CVE pull enabled but TRIVY_URL and TRIVY_TOKEN are not configured")
	}
	// Trivy server has no standard "list all image CVEs" endpoint; use webhooks for ingestion.
	return []Finding{}, nil
}
