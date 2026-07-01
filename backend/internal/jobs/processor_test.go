package jobs

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

func TestUnwrapTaskEnvelope(t *testing.T) {
	body := []byte(`{"hello":"world"}`)
	id := uuid.New()
	envelope, _ := json.Marshal(taskEnvelope{EventID: id.String(), Body: body})
	_, gotBody, gotID := unwrapTask(envelope)
	if gotID != id {
		t.Fatalf("event id: got %v want %v", gotID, id)
	}
	if string(gotBody) != string(body) {
		t.Fatalf("body mismatch: %s", gotBody)
	}
}

func TestUnwrapTaskRawPayload(t *testing.T) {
	raw := []byte(`{"legacy":true}`)
	_, body, id := unwrapTask(raw)
	if id != uuid.Nil {
		t.Fatal("expected nil event id for raw payload")
	}
	if string(body) != string(raw) {
		t.Fatal("expected raw body passthrough")
	}
}

func TestReportHashUsesDedupKey(t *testing.T) {
	body := []byte(`{"same":"payload"}`)
	h1 := reportHash(body, "dedup-a")
	h2 := reportHash(body, "dedup-b")
	h3 := reportHash(body, "")
	h4 := reportHash(body, "")
	if h1 == h2 {
		t.Fatal("dedup keys should produce different hashes")
	}
	if h3 != h4 {
		t.Fatal("empty dedup key should hash full payload consistently")
	}
}

func TestSplitImage(t *testing.T) {
	tests := []struct {
		artifact, wantName, wantTag string
	}{
		{"registry.io/app/service:v1.2.3", "registry.io/app/service", "v1.2.3"},
		{"myapp", "myapp", "latest"},
		{"host/repo:port", "host/repo", "port"},
	}
	for _, tc := range tests {
		name, tag := splitImage(tc.artifact)
		if name != tc.wantName || tag != tc.wantTag {
			t.Errorf("splitImage(%q) = (%q, %q), want (%q, %q)", tc.artifact, name, tag, tc.wantName, tc.wantTag)
		}
	}
}

func TestNormalizeSeverity(t *testing.T) {
	tests := map[string]string{
		"CRITICAL": "critical",
		"HIGH":     "high",
		"MEDIUM":   "medium",
		"LOW":      "low",
		"UNKNOWN":  "unknown",
		"":         "unknown",
	}
	for in, want := range tests {
		if got := normalizeSeverity(in); got != want {
			t.Errorf("normalizeSeverity(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestPayloadHashStable(t *testing.T) {
	a := payloadHash([]byte("payload"))
	b := payloadHash([]byte("payload"))
	if a != b || len(a) != 64 {
		t.Fatalf("unexpected hash: %q", a)
	}
}
