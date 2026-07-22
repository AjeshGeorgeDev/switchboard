package notifications

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/switchboard/switchboard/internal/db"
	"github.com/switchboard/switchboard/internal/settings"
	"github.com/switchboard/switchboard/internal/testutil"
)

func TestSendOutboundLogsSkippedWhenSMTPMissing(t *testing.T) {
	queries, pool := testutil.QueriesAndPool(t)
	ctx := context.Background()

	err := SendOutbound(ctx, queries, settings.SMTPConfig{}, []MailRecipient{
		RecipientForEmail("vp@example.com"),
	}, OutboundOptions{
		EventType: "smtp_test",
		Subject:   "Skip test",
		PlainBody: "no smtp",
	})
	if err != nil {
		t.Fatal(err)
	}

	rows, err := queries.ListEmailOutboundLog(ctx, db.ListEmailOutboundLogParams{
		Column1: "smtp_test",
		Limit:   10,
		Offset:  0,
	})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	var logID uuid.UUID
	for _, row := range rows {
		if row.Subject == "Skip test" && row.Status == "skipped" {
			found = true
			logID = row.ID
			break
		}
	}
	if !found {
		t.Fatal("expected skipped log row")
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM email_outbound_log WHERE id = $1`, logID)
	})

	recs, err := queries.ListEmailOutboundRecipientsByLogIDs(ctx, []uuid.UUID{logID})
	if err != nil {
		t.Fatal(err)
	}
	if len(recs) != 1 || recs[0].Email != "vp@example.com" || recs[0].Status != "skipped" {
		t.Fatalf("recipients: %+v", recs)
	}
}

func TestUsersToMailRecipientsDedupes(t *testing.T) {
	id := uuid.New()
	users := []db.User{
		{ID: id, Email: "a@example.com"},
		{ID: uuid.New(), Email: "A@example.com"},
		{ID: uuid.New(), Email: ""},
	}
	got := usersToMailRecipients(users)
	if len(got) != 1 || !strings.EqualFold(got[0].Email, "a@example.com") {
		t.Fatalf("got %+v", got)
	}
}
