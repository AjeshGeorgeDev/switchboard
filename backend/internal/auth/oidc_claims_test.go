package auth

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

func TestExtractOIDCIdentityDefaults(t *testing.T) {
	claims := map[string]any{
		"sub":   "abc-123",
		"email": "a@example.com",
		"name":  "Ada",
		"groups": []any{"admins", "readers"},
	}
	id, err := extractOIDCIdentity(claims, "", "", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if id.Subject != "abc-123" || id.Email != "a@example.com" || id.DisplayName != "Ada" {
		t.Fatalf("unexpected identity: %+v", id)
	}
	if len(id.Groups) != 2 || id.Groups[0] != "admins" {
		t.Fatalf("unexpected groups: %+v", id.Groups)
	}
}

func TestExtractOIDCIdentityCustomKeys(t *testing.T) {
	claims := map[string]any{
		"oid":                "subject-1",
		"preferred_username": "user@contoso.com",
		"displayName":        "Contoso User",
		"roles":              "switchboard-admins",
	}
	id, err := extractOIDCIdentity(claims, "oid", "preferred_username", "displayName", "roles")
	if err != nil {
		t.Fatal(err)
	}
	if id.Subject != "subject-1" || id.Email != "user@contoso.com" || id.DisplayName != "Contoso User" {
		t.Fatalf("unexpected identity: %+v", id)
	}
	if len(id.Groups) != 1 || id.Groups[0] != "switchboard-admins" {
		t.Fatalf("unexpected groups: %+v", id.Groups)
	}
}

func TestExtractOIDCIdentityMissingSubject(t *testing.T) {
	_, err := extractOIDCIdentity(map[string]any{"email": "a@b.c"}, "sub", "email", "name", "groups")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestClaimGroupsStringAndArray(t *testing.T) {
	if got := claimGroups(map[string]any{"g": "one"}, "g"); len(got) != 1 || got[0] != "one" {
		t.Fatalf("string claim: %+v", got)
	}
	if got := claimGroups(map[string]any{"g": []string{"a", "b"}}, "g"); len(got) != 2 {
		t.Fatalf("[]string claim: %+v", got)
	}
	if got := claimGroups(map[string]any{"g": []any{"x", 1}}, "g"); len(got) != 2 || got[1] != "1" {
		t.Fatalf("[]any claim: %+v", got)
	}
}

func TestMatchedRoleIDs(t *testing.T) {
	admin := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	viewer := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	mappings := []GroupRoleMapping{
		{Group: "admins", RoleID: admin.String()},
		{Group: "viewers", RoleID: viewer.String()},
		{Group: "admins", RoleID: admin.String()}, // dup
		{Group: "bad", RoleID: "not-a-uuid"},
	}
	got := matchedRoleIDs([]string{"admins", "other"}, mappings)
	if len(got) != 1 || got[0] != admin {
		t.Fatalf("matched: %+v", got)
	}
	if got := matchedRoleIDs(nil, mappings); got != nil {
		t.Fatalf("empty groups should return nil, got %+v", got)
	}
}

func TestParseAndNormalizeMappings(t *testing.T) {
	role := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	raw := json.RawMessage(`[{"group":" g1 ","role_id":"` + role.String() + `"},{"group":"","role_id":"` + role.String() + `"}]`)
	parsed := parseGroupRoleMappings(raw)
	if len(parsed) != 1 || parsed[0].Group != "g1" {
		t.Fatalf("parse: %+v", parsed)
	}
	normalized := NormalizeGroupRoleMappingsJSON(raw)
	var roundTrip []GroupRoleMapping
	if err := json.Unmarshal(normalized, &roundTrip); err != nil {
		t.Fatal(err)
	}
	if len(roundTrip) != 1 {
		t.Fatalf("normalize: %s", normalized)
	}
}
