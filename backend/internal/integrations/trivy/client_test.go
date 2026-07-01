package trivy

import (
	"context"
	"testing"

	"github.com/switchboard/switchboard/internal/config"
)

func TestFetchAllCVEsDisabled(t *testing.T) {
	c := NewClient(config.Config{CVEPullEnabled: false})
	findings, err := c.FetchAllCVEs(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if findings != nil {
		t.Fatalf("expected nil slice, got %v", findings)
	}
}

func TestFetchAllCVEsEnabledWithoutConfig(t *testing.T) {
	c := NewClient(config.Config{CVEPullEnabled: true})
	_, err := c.FetchAllCVEs(context.Background())
	if err == nil {
		t.Fatal("expected error when TRIVY_URL/TRIVY_TOKEN unset")
	}
}

func TestFetchAllCVEsEnabledWithConfig(t *testing.T) {
	c := NewClient(config.Config{
		CVEPullEnabled: true,
		TrivyURL:       "http://trivy.local",
		TrivyToken:     "token",
	})
	findings, err := c.FetchAllCVEs(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 0 {
		t.Fatalf("expected empty findings placeholder, got %d", len(findings))
	}
}
