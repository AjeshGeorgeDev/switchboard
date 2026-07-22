package notifications

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/switchboard/switchboard/internal/config"
	"github.com/switchboard/switchboard/internal/db"
	"github.com/switchboard/switchboard/internal/settings"
)

type Event struct {
	Type     string
	Title    string
	Body     string
	Severity string
	LinkURL  string
}

type Service struct {
	queries *db.Queries
	cfg     config.Config
	client  *asynq.Client
}

func NewService(queries *db.Queries, cfg config.Config, client *asynq.Client) *Service {
	return &Service{queries: queries, cfg: cfg, client: client}
}

func (s *Service) Notify(ctx context.Context, event Event) error {
	users, err := s.eligibleUsers(ctx, event.Type)
	if err != nil {
		return err
	}

	for _, user := range users {
		if s.channelEnabled(ctx, user.ID, "in_app", event.Type) {
			_, _ = s.queries.CreateNotification(ctx, db.CreateNotificationParams{
				UserID:    user.ID,
				Title:     event.Title,
				Body:      event.Body,
				EventType: event.Type,
				Severity:  event.Severity,
				LinkUrl:   pgtype.Text{String: event.LinkURL, Valid: event.LinkURL != ""},
			})
		}
	}

	if s.anyChannelEnabled(ctx, users, "teams", event.Type) {
		payload, _ := json.Marshal(event)
		_, _ = s.client.Enqueue(asynq.NewTask("notify:teams", payload))
	}

	if event.Type == "weekly_digest" || event.Type == "critical_cve" {
		if s.anyChannelEnabled(ctx, users, "email", event.Type) {
			payload, _ := json.Marshal(event)
			_, _ = s.client.Enqueue(asynq.NewTask("notify:email", payload))
		}
	}

	return nil
}

func (s *Service) eligibleUsers(ctx context.Context, eventType string) ([]db.User, error) {
	switch eventType {
	case "weekly_digest", "critical_cve", "deployment_report":
		return s.queries.GetUsersByRoleName(ctx, "security-team")
	default:
		users, err := s.queries.ListUsers(ctx)
		return users, err
	}
}

func (s *Service) channelEnabled(ctx context.Context, userID uuid.UUID, channel, eventType string) bool {
	prefs, err := s.queries.GetNotificationPreferences(ctx, userID)
	if err != nil || len(prefs) == 0 {
		return defaultEnabled(channel, eventType)
	}
	for _, p := range prefs {
		if string(p.Channel) == channel && string(p.EventType) == eventType {
			return p.Enabled
		}
	}
	return defaultEnabled(channel, eventType)
}

func defaultEnabled(channel, eventType string) bool {
	if channel == "email" && eventType == "deployment_report" {
		return false
	}
	return true
}

func (s *Service) anyChannelEnabled(ctx context.Context, users []db.User, channel, eventType string) bool {
	for _, u := range users {
		prefs, _ := s.queries.GetNotificationPreferences(ctx, u.ID)
		if len(prefs) == 0 {
			if defaultEnabled(channel, eventType) {
				return true
			}
			continue
		}
		for _, p := range prefs {
			if string(p.Channel) == channel && string(p.EventType) == eventType && p.Enabled {
				return true
			}
		}
	}
	return false
}

func (s *Service) SendTeamsPayload(ctx context.Context, payload []byte) error {
	var event Event
	if err := json.Unmarshal(payload, &event); err != nil {
		return err
	}
	configs, err := s.queries.ListActiveTeamsWebhookConfigs(ctx)
	if err != nil {
		return err
	}
	card := teamsCard(event)
	body, _ := json.Marshal(card)
	for _, cfg := range configs {
		if !containsEvent(cfg.EventTypes, event.Type) {
			continue
		}
		req, _ := http.NewRequestWithContext(ctx, http.MethodPost, cfg.WebhookUrl, strings.NewReader(string(body)))
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		resp.Body.Close()
	}
	return nil
}

func (s *Service) SendEmailPayload(ctx context.Context, payload []byte) error {
	var event Event
	if err := json.Unmarshal(payload, &event); err != nil {
		return err
	}
	smtp := settings.ResolveSMTP(ctx, s.queries, s.cfg)
	if !smtp.Configured() {
		return nil
	}
	users, _ := s.queries.GetUsersByRoleName(ctx, "security-team")
	return sendSMTP(smtp, users, event)
}

func teamsCard(event Event) map[string]interface{} {
	return map[string]interface{}{
		"type":    "message",
		"attachments": []map[string]interface{}{
			{
				"contentType": "application/vnd.microsoft.card.adaptive",
				"content": map[string]interface{}{
					"type":    "AdaptiveCard",
					"version": "1.4",
					"body": []map[string]interface{}{
						{"type": "TextBlock", "text": event.Title, "weight": "Bolder", "size": "Medium"},
						{"type": "TextBlock", "text": event.Body},
					},
					"actions": []map[string]interface{}{
						{"type": "Action.OpenUrl", "title": "View in Dashboard", "url": event.LinkURL},
					},
				},
			},
		},
	}
}

func containsEvent(types []string, eventType string) bool {
	for _, t := range types {
		if t == eventType || t == "*" {
			return true
		}
	}
	return len(types) == 0
}
