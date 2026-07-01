package config

import "testing"

func TestLoadCVEPullDisabledByDefault(t *testing.T) {
	t.Setenv("CVE_PULL_ENABLED", "")
	cfg := Load()
	if cfg.CVEPullEnabled {
		t.Fatal("CVE_PULL_ENABLED should default to false")
	}
}

func TestLoadCVEPullEnabled(t *testing.T) {
	t.Setenv("CVE_PULL_ENABLED", "true")
	cfg := Load()
	if !cfg.CVEPullEnabled {
		t.Fatal("expected CVE pull enabled")
	}
}
