package settings

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/switchboard/switchboard/internal/db"
)

var defaultEmailRoles = []string{"security-team"}

// EmailRecipientRoles returns configured role names for a security email event type.
func EmailRecipientRoles(ctx context.Context, q *db.Queries, eventType string) []string {
	key := ""
	switch eventType {
	case "weekly_digest":
		key = KeyEmailRolesWeeklyDigest
	case "critical_cve":
		key = KeyEmailRolesCriticalCVE
	default:
		return append([]string{}, defaultEmailRoles...)
	}
	raw := settingOr(ctx, q, key, "")
	if raw == "" {
		return append([]string{}, defaultEmailRoles...)
	}
	var roles []string
	if err := json.Unmarshal([]byte(raw), &roles); err != nil {
		return append([]string{}, defaultEmailRoles...)
	}
	out := make([]string, 0, len(roles))
	seen := map[string]struct{}{}
	for _, r := range roles {
		r = strings.TrimSpace(r)
		if r == "" {
			continue
		}
		if _, ok := seen[r]; ok {
			continue
		}
		seen[r] = struct{}{}
		out = append(out, r)
	}
	if len(out) == 0 {
		return append([]string{}, defaultEmailRoles...)
	}
	return out
}

// SaveEmailRecipientRoles persists role lists for digest and critical CVE emails.
func SaveEmailRecipientRoles(ctx context.Context, q *db.Queries, digestRoles, criticalRoles []string) error {
	digestJSON, err := json.Marshal(normalizeRoleList(digestRoles))
	if err != nil {
		return err
	}
	criticalJSON, err := json.Marshal(normalizeRoleList(criticalRoles))
	if err != nil {
		return err
	}
	if err := upsertSetting(ctx, q, KeyEmailRolesWeeklyDigest, string(digestJSON)); err != nil {
		return err
	}
	return upsertSetting(ctx, q, KeyEmailRolesCriticalCVE, string(criticalJSON))
}

func normalizeRoleList(roles []string) []string {
	out := make([]string, 0, len(roles))
	seen := map[string]struct{}{}
	for _, r := range roles {
		r = strings.TrimSpace(r)
		if r == "" {
			continue
		}
		if _, ok := seen[r]; ok {
			continue
		}
		seen[r] = struct{}{}
		out = append(out, r)
	}
	if len(out) == 0 {
		return append([]string{}, defaultEmailRoles...)
	}
	return out
}
