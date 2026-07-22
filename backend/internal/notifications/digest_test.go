package notifications

import (
	"strings"
	"testing"
)

func TestRenderWeeklyDigestHTML(t *testing.T) {
	days := 12
	html, err := RenderWeeklyDigestHTML(DigestEmailData{
		Title:           "Weekly security digest",
		Intro:           "Harbor vulnerability posture from Switchboard.",
		OverviewURL:     "http://localhost:5173/security",
		Critical:        3,
		High:            7,
		NewThisWeek:     2,
		FixableCritical: 1,
		AgingLt7d:       1,
		AgingGt30d:      2,
		TopImages: []DigestImageRow{
			{Name: "library/nginx", Critical: 2, High: 1, OldestDays: &days},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Critical", "library/nginx", "/security", "Fixable criticals"} {
		if !strings.Contains(html, want) {
			t.Fatalf("expected %q in html:\n%s", want, html)
		}
	}
}
