package notifications

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/switchboard/switchboard/internal/db"
	"github.com/switchboard/switchboard/internal/settings"
	"gopkg.in/gomail.v2"
)

type MailRecipient struct {
	Email   string
	UserID  uuid.UUID
	HasUser bool
}

type OutboundOptions struct {
	EventType   string
	Subject     string
	HTMLBody    string
	PlainBody   string
	TriggeredBy *uuid.UUID
}

func truncatePreview(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func pgUUID(id *uuid.UUID) pgtype.UUID {
	if id == nil || *id == uuid.Nil {
		return pgtype.UUID{}
	}
	return pgtype.UUID{Bytes: *id, Valid: true}
}

func pgUserID(id uuid.UUID, ok bool) pgtype.UUID {
	if !ok || id == uuid.Nil {
		return pgtype.UUID{}
	}
	return pgtype.UUID{Bytes: id, Valid: true}
}

func textErr(err error) pgtype.Text {
	if err == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: err.Error(), Valid: true}
}

// deliverSMTP sends one message per recipient and writes outbound log rows.
func deliverSMTP(ctx context.Context, q *db.Queries, smtp settings.SMTPConfig, recipients []MailRecipient, opt OutboundOptions) error {
	preview := truncatePreview(opt.PlainBody, 500)
	if preview == "" {
		preview = truncatePreview(stripTags(opt.HTMLBody), 500)
	}

	if !smtp.Configured() {
		_, _ = writeOutboundLog(ctx, q, opt, preview, "skipped", fmt.Errorf("SMTP not configured"), recipients, true)
		return nil
	}
	if len(recipients) == 0 {
		_, _ = writeOutboundLog(ctx, q, opt, preview, "skipped", fmt.Errorf("no recipients"), nil, false)
		return nil
	}

	d := gomail.NewDialer(smtp.Host, smtp.Port, smtp.User, smtp.Pass)
	var firstErr error
	results := make([]struct {
		r   MailRecipient
		err error
	}, 0, len(recipients))

	for _, r := range recipients {
		m := gomail.NewMessage()
		m.SetHeader("From", smtp.From)
		m.SetHeader("To", r.Email)
		m.SetHeader("Subject", opt.Subject)
		if opt.HTMLBody != "" {
			m.SetBody("text/html", opt.HTMLBody)
		} else {
			m.SetBody("text/plain", opt.PlainBody)
		}
		err := d.DialAndSend(m)
		results = append(results, struct {
			r   MailRecipient
			err error
		}{r: r, err: err})
		if err != nil && firstErr == nil {
			firstErr = err
		}
	}

	status := "sent"
	var logErr error
	if firstErr != nil {
		allFailed := true
		for _, res := range results {
			if res.err == nil {
				allFailed = false
				break
			}
		}
		if allFailed {
			status = "failed"
			logErr = firstErr
		} else {
			status = "sent"
			logErr = firstErr
		}
	}

	logRow, err := writeOutboundLog(ctx, q, opt, preview, status, logErr, nil, false)
	if err != nil {
		if firstErr != nil {
			return firstErr
		}
		return err
	}
	for _, res := range results {
		recStatus := "sent"
		var recErr error
		if res.err != nil {
			recStatus = "failed"
			recErr = res.err
		}
		_, _ = q.CreateEmailOutboundRecipient(ctx, db.CreateEmailOutboundRecipientParams{
			LogID:        logRow.ID,
			Email:        res.r.Email,
			UserID:       pgUserID(res.r.UserID, res.r.HasUser),
			Status:       recStatus,
			ErrorMessage: textErr(recErr),
		})
	}
	return firstErr
}

func writeOutboundLog(
	ctx context.Context,
	q *db.Queries,
	opt OutboundOptions,
	preview, status string,
	logErr error,
	skippedRecipients []MailRecipient,
	writeSkippedRecipients bool,
) (db.EmailOutboundLog, error) {
	row, err := q.CreateEmailOutboundLog(ctx, db.CreateEmailOutboundLogParams{
		EventType:    opt.EventType,
		Subject:      opt.Subject,
		BodyPreview:  preview,
		Status:       status,
		ErrorMessage: textErr(logErr),
		TriggeredBy:  pgUUID(opt.TriggeredBy),
	})
	if err != nil {
		return row, err
	}
	if writeSkippedRecipients {
		for _, r := range skippedRecipients {
			_, _ = q.CreateEmailOutboundRecipient(ctx, db.CreateEmailOutboundRecipientParams{
				LogID:        row.ID,
				Email:        r.Email,
				UserID:       pgUserID(r.UserID, r.HasUser),
				Status:       "skipped",
				ErrorMessage: textErr(logErr),
			})
		}
	}
	return row, nil
}

func stripTags(html string) string {
	var b strings.Builder
	inTag := false
	for _, r := range html {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case !inTag:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// SendOutbound sends mail to recipients and records an outbound log entry.
func SendOutbound(ctx context.Context, q *db.Queries, smtp settings.SMTPConfig, recipients []MailRecipient, opt OutboundOptions) error {
	return deliverSMTP(ctx, q, smtp, recipients, opt)
}

// RecipientForEmail builds a non-user mail recipient (e.g. invitee).
func RecipientForEmail(email string) MailRecipient {
	return MailRecipient{Email: strings.TrimSpace(email)}
}

// RecipientForUser builds a mail recipient from a user record.
func RecipientForUser(userID uuid.UUID, email string) MailRecipient {
	return MailRecipient{Email: strings.TrimSpace(email), UserID: userID, HasUser: userID != uuid.Nil}
}

func usersToMailRecipients(users []db.User) []MailRecipient {
	out := make([]MailRecipient, 0, len(users))
	seen := map[string]struct{}{}
	for _, u := range users {
		email := strings.TrimSpace(u.Email)
		if email == "" {
			continue
		}
		key := strings.ToLower(email)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, MailRecipient{Email: email, UserID: u.ID, HasUser: true})
	}
	return out
}
