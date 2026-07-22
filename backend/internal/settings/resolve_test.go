package settings

import (
	"context"
	"testing"

	"github.com/switchboard/switchboard/internal/config"
)

func TestResolveHarborFallsBackToEnv(t *testing.T) {
	cfg := ResolveHarbor(context.Background(), nil, config.Config{
		HarborURL:           "https://harbor.example.com",
		HarborUser:          "robot$lib",
		HarborToken:         "secret",
		HarborWebhookSecret: "hmac",
	})
	if !cfg.APIConfigured() {
		t.Fatal("expected API configured from env")
	}
	if cfg.URL != "https://harbor.example.com" || cfg.WebhookSecret != "hmac" {
		t.Fatalf("unexpected config: %+v", cfg)
	}
}

func TestHarborAPIConfigured(t *testing.T) {
	if (HarborConfig{URL: "https://h", Token: "secret"}).APIConfigured() {
		t.Fatal("token without user should not configure unless user:pass")
	}
	if !(HarborConfig{URL: "https://h", Token: "user:pass"}).APIConfigured() {
		t.Fatal("combined token should configure")
	}
	if !(HarborConfig{URL: "https://h", User: "u", Token: "t"}).APIConfigured() {
		t.Fatal("user+token should configure")
	}
}

func TestResolveSMTPFallsBackToEnv(t *testing.T) {
	cfg := ResolveSMTP(context.Background(), nil, config.Config{
		SMTPHost: "smtp.example.com",
		SMTPPort: 465,
		SMTPUser: "u",
		SMTPPass: "p",
		SMTPFrom: "from@example.com",
	})
	if !cfg.Configured() || cfg.Port != 465 || cfg.From != "from@example.com" {
		t.Fatalf("unexpected smtp: %+v", cfg)
	}
}
