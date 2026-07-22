package settings

import (
	"context"
	"strconv"
	"strings"

	"github.com/switchboard/switchboard/internal/config"
	"github.com/switchboard/switchboard/internal/db"
)

const (
	KeyHarborURL           = "harbor.url"
	KeyHarborUser          = "harbor.user"
	KeyHarborToken         = "harbor.token"
	KeyHarborWebhookSecret = "harbor.webhook_secret"

	KeySMTPHost = "smtp.host"
	KeySMTPPort = "smtp.port"
	KeySMTPUser = "smtp.user"
	KeySMTPPass = "smtp.pass"
	KeySMTPFrom = "smtp.from"
)

// HarborConfig is the effective Harbor integration config (DB overrides env).
type HarborConfig struct {
	URL           string
	User          string
	Token         string
	WebhookSecret string
}

func (h HarborConfig) APIConfigured() bool {
	if strings.TrimSpace(h.URL) == "" || strings.TrimSpace(h.Token) == "" {
		return false
	}
	if strings.TrimSpace(h.User) != "" {
		return true
	}
	return strings.Contains(h.Token, ":")
}

// SMTPConfig is the effective SMTP config (DB overrides env).
type SMTPConfig struct {
	Host string
	Port int
	User string
	Pass string
	From string
}

func (s SMTPConfig) Configured() bool {
	return strings.TrimSpace(s.Host) != ""
}

// ResolveHarbor loads Harbor settings from app_settings, falling back to env.
func ResolveHarbor(ctx context.Context, q *db.Queries, env config.Config) HarborConfig {
	return HarborConfig{
		URL:           settingOr(ctx, q, KeyHarborURL, env.HarborURL),
		User:          settingOr(ctx, q, KeyHarborUser, env.HarborUser),
		Token:         settingOr(ctx, q, KeyHarborToken, env.HarborToken),
		WebhookSecret: settingOr(ctx, q, KeyHarborWebhookSecret, env.HarborWebhookSecret),
	}
}

// ResolveSMTP loads SMTP settings from app_settings, falling back to env.
func ResolveSMTP(ctx context.Context, q *db.Queries, env config.Config) SMTPConfig {
	port := env.SMTPPort
	if raw := settingOr(ctx, q, KeySMTPPort, ""); raw != "" {
		if p, err := strconv.Atoi(raw); err == nil && p > 0 {
			port = p
		}
	}
	return SMTPConfig{
		Host: settingOr(ctx, q, KeySMTPHost, env.SMTPHost),
		Port: port,
		User: settingOr(ctx, q, KeySMTPUser, env.SMTPUser),
		Pass: settingOr(ctx, q, KeySMTPPass, env.SMTPPass),
		From: settingOr(ctx, q, KeySMTPFrom, env.SMTPFrom),
	}
}

func settingOr(ctx context.Context, q *db.Queries, key, fallback string) string {
	if q == nil {
		return strings.TrimSpace(fallback)
	}
	row, err := q.GetAppSetting(ctx, key)
	if err == nil && strings.TrimSpace(row.Value) != "" {
		return strings.TrimSpace(row.Value)
	}
	return strings.TrimSpace(fallback)
}

func upsertSetting(ctx context.Context, q *db.Queries, key, value string) error {
	_, err := q.UpsertAppSetting(ctx, db.UpsertAppSettingParams{Key: key, Value: value})
	return err
}

func clearSetting(ctx context.Context, q *db.Queries, key string) error {
	return upsertSetting(ctx, q, key, "")
}
