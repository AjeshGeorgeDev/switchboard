package webhooks

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/switchboard/switchboard/internal/config"
)

func signBody(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

func TestVerifySecret(t *testing.T) {
	body := []byte(`{"event":"test"}`)
	secret := "super-secret"
	sig := signBody(secret, body)

	if !verifySecret(body, sig, secret) {
		t.Fatal("expected valid signature")
	}
	if verifySecret(body, "bad", secret) {
		t.Fatal("expected invalid signature to fail")
	}
	if verifySecret(body, "", secret) {
		t.Fatal("expected missing signature to fail when secret set")
	}
	if !verifySecret(body, "", "") {
		t.Fatal("expected open mode when secret unset")
	}
}

func TestPaginationDefaults(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/?limit=0&offset=10", nil)
	limit, offset := pagination(req)
	if limit != 50 || offset != 10 {
		t.Fatalf("got limit=%d offset=%d", limit, offset)
	}
}

func TestPaginationInvalidLimitUsesDefault(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/?limit=999", nil)
	limit, _ := pagination(req)
	if limit != 50 {
		t.Fatalf("expected default 50, got %d", limit)
	}
}

func TestEndpointsHandler(t *testing.T) {
	h := NewHandler(nil, config.Config{
		AppBaseURL:          "http://localhost:8080/",
		HarborWebhookSecret: "h",
		HarborURL:           "https://harbor.example.com",
		HarborToken:         "robot:secret",
		CVEPullEnabled:      true,
		CVEPullCron:         "0 6 * * 0",
		TrivyURL:            "http://trivy",
		TrivyToken:          "token",
	}, nil)

	rec := httptest.NewRecorder()
	h.Endpoints(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	var out map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if out["harbor_url"] != "http://localhost:8080/webhooks/harbor" {
		t.Fatalf("harbor_url: %v", out["harbor_url"])
	}
	if out["harbor_secret_configured"] != true {
		t.Fatal("expected harbor secret configured")
	}
	if out["harbor_api_configured"] != true {
		t.Fatal("expected harbor API configured")
	}
	if out["cve_pull_enabled"] != true {
		t.Fatal("expected cve pull enabled flag")
	}
}

func TestHarborRejectsInvalidSignature(t *testing.T) {
	h := NewHandler(nil, config.Config{HarborWebhookSecret: "secret"}, nil)
	body := []byte(`{}`)
	req := httptest.NewRequest(http.MethodPost, "/webhooks/harbor", strings.NewReader(string(body)))
	req.Header.Set("X-Webhook-Signature", "invalid")
	rec := httptest.NewRecorder()
	h.Harbor(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status %d, want 401", rec.Code)
	}
}
