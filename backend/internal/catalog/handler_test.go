package catalog

import (
	"encoding/json"
	"testing"
)

func TestAppRequestDecodesIsPublic(t *testing.T) {
	body := `{
		"name":"Google",
		"description":"",
		"icon_url":"",
		"access_type":"url",
		"target_host":"https://google.com",
		"target_port":null,
		"is_active":true,
		"is_public":true,
		"sort_order":0,
		"role_ids":[]
	}`

	var req appRequest
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !req.IsActive {
		t.Fatal("expected is_active true")
	}
	if !req.IsPublic {
		t.Fatal("expected is_public true")
	}
}
