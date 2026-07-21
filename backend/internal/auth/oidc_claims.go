package auth

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type GroupRoleMapping struct {
	Group  string `json:"group"`
	RoleID string `json:"role_id"`
}

type oidcIdentity struct {
	Subject     string
	Email       string
	DisplayName string
	Groups      []string
}

func claimKeyOrDefault(configured, fallback string) string {
	key := strings.TrimSpace(configured)
	if key == "" {
		return fallback
	}
	return key
}

func claimString(claims map[string]any, key string) string {
	if key == "" || claims == nil {
		return ""
	}
	v, ok := claims[key]
	if !ok || v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t)
	case fmt.Stringer:
		return strings.TrimSpace(t.String())
	default:
		return strings.TrimSpace(fmt.Sprint(t))
	}
}

func claimGroups(claims map[string]any, key string) []string {
	if key == "" || claims == nil {
		return nil
	}
	v, ok := claims[key]
	if !ok || v == nil {
		return nil
	}
	switch t := v.(type) {
	case string:
		s := strings.TrimSpace(t)
		if s == "" {
			return nil
		}
		return []string{s}
	case []any:
		out := make([]string, 0, len(t))
		for _, item := range t {
			s := strings.TrimSpace(fmt.Sprint(item))
			if s != "" && s != "<nil>" {
				out = append(out, s)
			}
		}
		return out
	case []string:
		out := make([]string, 0, len(t))
		for _, item := range t {
			s := strings.TrimSpace(item)
			if s != "" {
				out = append(out, s)
			}
		}
		return out
	default:
		return nil
	}
}

func extractOIDCIdentity(claims map[string]any, claimSubject, claimEmail, claimName, claimGroupsKey string) (oidcIdentity, error) {
	subjectKey := claimKeyOrDefault(claimSubject, "sub")
	emailKey := claimKeyOrDefault(claimEmail, "email")
	nameKey := claimKeyOrDefault(claimName, "name")
	groupsKey := claimKeyOrDefault(claimGroupsKey, "groups")

	id := oidcIdentity{
		Subject:     claimString(claims, subjectKey),
		Email:       claimString(claims, emailKey),
		DisplayName: claimString(claims, nameKey),
		Groups:      claimGroups(claims, groupsKey),
	}
	if id.Subject == "" {
		return oidcIdentity{}, fmt.Errorf("missing subject claim %q", subjectKey)
	}
	return id, nil
}

func parseGroupRoleMappings(raw []byte) []GroupRoleMapping {
	if len(raw) == 0 {
		return nil
	}
	var mappings []GroupRoleMapping
	if err := json.Unmarshal(raw, &mappings); err != nil {
		return nil
	}
	out := make([]GroupRoleMapping, 0, len(mappings))
	for _, m := range mappings {
		group := strings.TrimSpace(m.Group)
		roleID := strings.TrimSpace(m.RoleID)
		if group == "" || roleID == "" {
			continue
		}
		if _, err := uuid.Parse(roleID); err != nil {
			continue
		}
		out = append(out, GroupRoleMapping{Group: group, RoleID: roleID})
	}
	return out
}

// matchedRoleIDs returns role UUIDs for groups that appear in both the token and mappings.
func matchedRoleIDs(groups []string, mappings []GroupRoleMapping) []uuid.UUID {
	if len(groups) == 0 || len(mappings) == 0 {
		return nil
	}
	groupSet := make(map[string]struct{}, len(groups))
	for _, g := range groups {
		groupSet[g] = struct{}{}
	}
	seen := make(map[uuid.UUID]struct{})
	var roles []uuid.UUID
	for _, m := range mappings {
		if _, ok := groupSet[m.Group]; !ok {
			continue
		}
		id, err := uuid.Parse(m.RoleID)
		if err != nil {
			continue
		}
		if _, dup := seen[id]; dup {
			continue
		}
		seen[id] = struct{}{}
		roles = append(roles, id)
	}
	return roles
}

func NormalizeGroupRoleMappingsJSON(raw json.RawMessage) []byte {
	mappings := parseGroupRoleMappings(raw)
	if mappings == nil {
		mappings = []GroupRoleMapping{}
	}
	b, err := json.Marshal(mappings)
	if err != nil {
		return []byte("[]")
	}
	return b
}
