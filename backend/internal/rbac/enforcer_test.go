package rbac

import "testing"

func TestHasPermission(t *testing.T) {
	e := &Enforcer{
		policies: []policy{
			{Role: "admin", Object: "*", Action: "*"},
			{Role: "security-team", Object: "security", Action: "read"},
			{Role: "viewer", Object: "catalog", Action: "read"},
		},
	}

	tests := []struct {
		roles      []string
		obj, act   string
		wantAllow  bool
	}{
		{[]string{"admin"}, "anything", "write", true},
		{[]string{"viewer"}, "catalog", "read", true},
		{[]string{"viewer"}, "security", "read", false},
		{[]string{"security-team"}, "security", "read", true},
		{[]string{"security-team"}, "admin", "*", false},
		{[]string{}, "catalog", "read", false},
		{[]string{"viewer"}, "catalog", "write", false},
	}
	for _, tc := range tests {
		got := e.HasPermission(tc.roles, tc.obj, tc.act)
		if got != tc.wantAllow {
			t.Errorf("HasPermission(%v, %q, %q) = %v, want %v", tc.roles, tc.obj, tc.act, got, tc.wantAllow)
		}
	}
}

func TestWildcardAction(t *testing.T) {
	e := &Enforcer{policies: []policy{{Role: "admin", Object: "admin", Action: "*"}}}
	if !e.HasPermission([]string{"admin"}, "admin", "delete") {
		t.Fatal("expected wildcard action to allow delete")
	}
}
