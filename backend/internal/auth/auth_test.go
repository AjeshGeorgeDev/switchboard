package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/switchboard/switchboard/internal/config"
)

func TestNormalizeEmail(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"  Admin@Example.COM ", "admin@example.com"},
		{"user@firm.io", "user@firm.io"},
		{"", ""},
	}
	for _, tc := range tests {
		if got := NormalizeEmail(tc.in); got != tc.want {
			t.Errorf("NormalizeEmail(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestEmailAsUsername(t *testing.T) {
	if got := EmailAsUsername("  Me@Example.com "); got != "me@example.com" {
		t.Fatalf("got %q", got)
	}
}

func TestTokenRoundTrip(t *testing.T) {
	svc := NewTokenService(config.Config{
		JWTSecret:     "test-secret",
		JWTAccessTTL:  15 * time.Minute,
	})
	userID := uuid.New()
	token, _, err := svc.CreateAccessToken(userID)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := svc.ParseAccessToken(token)
	if err != nil {
		t.Fatal(err)
	}
	if parsed != userID {
		t.Fatalf("user id mismatch: got %s want %s", parsed, userID)
	}
}

func TestParseAccessTokenRejectsBadSecret(t *testing.T) {
	a := NewTokenService(config.Config{JWTSecret: "a", JWTAccessTTL: 15 * time.Minute})
	b := NewTokenService(config.Config{JWTSecret: "b", JWTAccessTTL: 15 * time.Minute})
	token, _, err := a.CreateAccessToken(uuid.New())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := b.ParseAccessToken(token); err == nil {
		t.Fatal("expected error for wrong secret")
	}
}

func TestHashTokenDeterministic(t *testing.T) {
	a := HashToken("refresh-token")
	b := HashToken("refresh-token")
	if a != b || a == "" {
		t.Fatalf("hash not stable: %q %q", a, b)
	}
}

func TestNewRefreshTokenUnique(t *testing.T) {
	a, err := NewRefreshToken()
	if err != nil {
		t.Fatal(err)
	}
	b, err := NewRefreshToken()
	if err != nil {
		t.Fatal(err)
	}
	if a == b || len(a) < 20 {
		t.Fatalf("expected unique non-trivial tokens, got %q and %q", a, b)
	}
}
