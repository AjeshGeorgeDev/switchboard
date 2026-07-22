package settings

import (
	"context"
	"testing"
)

func TestEmailRecipientRolesDefault(t *testing.T) {
	roles := EmailRecipientRoles(context.Background(), nil, "weekly_digest")
	if len(roles) != 1 || roles[0] != "security-team" {
		t.Fatalf("got %#v", roles)
	}
	roles = EmailRecipientRoles(context.Background(), nil, "critical_cve")
	if len(roles) != 1 || roles[0] != "security-team" {
		t.Fatalf("got %#v", roles)
	}
}

func TestNormalizeRoleList(t *testing.T) {
	got := normalizeRoleList([]string{" admin ", "security-team", "admin", ""})
	if len(got) != 2 || got[0] != "admin" || got[1] != "security-team" {
		t.Fatalf("got %#v", got)
	}
	got = normalizeRoleList(nil)
	if len(got) != 1 || got[0] != "security-team" {
		t.Fatalf("empty should default, got %#v", got)
	}
}
