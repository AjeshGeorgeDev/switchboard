package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port                   string
	DatabaseURL            string
	RedisURL               string
	JWTSecret              string
	JWTAccessTTL           time.Duration
	JWTSessionRefreshTTL   time.Duration
	JWTRememberRefreshTTL  time.Duration
	HarborURL              string
	HarborUser             string
	HarborToken            string
	TrivyURL               string
	TrivyToken             string
	HarborWebhookSecret    string
	TrivyWebhookSecret     string
	CVEPullCron            string
	CVEPullEnabled         bool
	SMTPHost               string
	SMTPPort               int
	SMTPUser               string
	SMTPPass               string
	SMTPFrom               string
	AppBaseURL             string
	NotificationRetention  int
	CVERetentionMonths     int
	WorkerMode             string
}

func Load() Config {
	smtpPort, _ := strconv.Atoi(getEnv("SMTP_PORT", "587"))
	notifRetention, _ := strconv.Atoi(getEnv("NOTIFICATION_RETENTION_DAYS", "90"))
	cveRetention, _ := strconv.Atoi(getEnv("CVE_RETENTION_MONTHS", "12"))
	rememberFallback := getEnv("JWT_REFRESH_TTL", "720h")

	return Config{
		Port:                   getEnv("PORT", "8080"),
		DatabaseURL:            getEnv("DATABASE_URL", "postgres://switchboard:switchboard@localhost:5432/switchboard?sslmode=disable"),
		RedisURL:               getEnv("REDIS_URL", "redis://localhost:6379/0"),
		JWTSecret:              getEnv("JWT_SECRET", "dev-secret-change-in-production"),
		JWTAccessTTL:           15 * time.Minute,
		JWTSessionRefreshTTL:   durationEnv("JWT_SESSION_REFRESH_TTL", "24h"),
		JWTRememberRefreshTTL:  durationEnv("JWT_REMEMBER_REFRESH_TTL", rememberFallback),
		HarborURL:              getEnv("HARBOR_URL", ""),
		HarborUser:             getEnv("HARBOR_USER", ""),
		HarborToken:            getEnv("HARBOR_TOKEN", ""),
		TrivyURL:               getEnv("TRIVY_URL", ""),
		TrivyToken:             getEnv("TRIVY_TOKEN", ""),
		HarborWebhookSecret:    getEnv("HARBOR_WEBHOOK_SECRET", ""),
		TrivyWebhookSecret:     getEnv("TRIVY_WEBHOOK_SECRET", ""),
		CVEPullCron:            getEnv("CVE_PULL_CRON", "0 6 * * 0"),
		CVEPullEnabled:         getEnv("CVE_PULL_ENABLED", "false") == "true",
		SMTPHost:               getEnv("SMTP_HOST", ""),
		SMTPPort:               smtpPort,
		SMTPUser:               getEnv("SMTP_USER", ""),
		SMTPPass:               getEnv("SMTP_PASS", ""),
		SMTPFrom:               getEnv("SMTP_FROM", ""),
		AppBaseURL:             getEnv("APP_BASE_URL", "http://localhost:8080"),
		NotificationRetention:  notifRetention,
		CVERetentionMonths:     cveRetention,
		WorkerMode:             getEnv("WORKER_MODE", "combined"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func durationEnv(key, fallback string) time.Duration {
	raw := getEnv(key, fallback)
	d, err := time.ParseDuration(raw)
	if err != nil {
		d, _ = time.ParseDuration(fallback)
	}
	return d
}
