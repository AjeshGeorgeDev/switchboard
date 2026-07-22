package notifications

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/switchboard/switchboard/internal/auth"
	"github.com/switchboard/switchboard/internal/config"
	"github.com/switchboard/switchboard/internal/db"
	"github.com/switchboard/switchboard/internal/settings"
	"gopkg.in/gomail.v2"
)

type Handler struct {
	queries *db.Queries
	cfg     config.Config
}

func NewHandler(queries *db.Queries, cfg config.Config) *Handler {
	return &Handler{queries: queries, cfg: cfg}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}
	limit := int32(10)
	if unread := r.URL.Query().Get("unread"); unread == "true" {
		items, err := h.queries.ListUnreadNotificationsForUser(r.Context(), db.ListUnreadNotificationsForUserParams{
			UserID: userID,
			Limit:  limit,
		})
		if err != nil {
			http.Error(w, `{"error":"server error"}`, http.StatusInternalServerError)
			return
		}
		count, _ := h.queries.CountUnreadNotifications(r.Context(), userID)
		auth.WriteJSON(w, http.StatusOK, map[string]interface{}{"items": items, "unread_count": count})
		return
	}
	items, err := h.queries.ListNotificationsForUser(r.Context(), db.ListNotificationsForUserParams{UserID: userID, Limit: limit})
	if err != nil {
		http.Error(w, `{"error":"server error"}`, http.StatusInternalServerError)
		return
	}
	count, _ := h.queries.CountUnreadNotifications(r.Context(), userID)
	auth.WriteJSON(w, http.StatusOK, map[string]interface{}{"items": items, "unread_count": count})
}

func (h *Handler) MarkRead(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}
	_ = h.queries.MarkNotificationRead(r.Context(), db.MarkNotificationReadParams{ID: id, UserID: userID})
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) MarkAllRead(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())
	_ = h.queries.MarkAllNotificationsRead(r.Context(), userID)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) GetPreferences(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())
	prefs, err := h.queries.GetNotificationPreferences(r.Context(), userID)
	if err != nil {
		http.Error(w, `{"error":"server error"}`, http.StatusInternalServerError)
		return
	}
	auth.WriteJSON(w, http.StatusOK, prefs)
}

func (h *Handler) UpdatePreferences(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())
	var body []struct {
		Channel   string `json:"channel"`
		EventType string `json:"event_type"`
		Enabled   bool   `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}
	for _, p := range body {
		_ = h.queries.UpsertNotificationPreference(r.Context(), db.UpsertNotificationPreferenceParams{
			UserID:    userID,
			Channel:   p.Channel,
			EventType: p.EventType,
			Enabled:   p.Enabled,
		})
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListTeamsWebhooks(w http.ResponseWriter, r *http.Request) {
	items, err := h.queries.ListTeamsWebhookConfigs(r.Context())
	if err != nil {
		http.Error(w, `{"error":"server error"}`, http.StatusInternalServerError)
		return
	}
	auth.WriteJSON(w, http.StatusOK, items)
}

func (h *Handler) CreateTeamsWebhook(w http.ResponseWriter, r *http.Request) {
	var body db.CreateTeamsWebhookConfigParams
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}
	item, err := h.queries.CreateTeamsWebhookConfig(r.Context(), body)
	if err != nil {
		http.Error(w, `{"error":"create failed"}`, http.StatusInternalServerError)
		return
	}
	auth.WriteJSON(w, http.StatusCreated, item)
}

func (h *Handler) UpdateTeamsWebhook(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}
	var body struct {
		Name       string   `json:"name"`
		WebhookUrl string   `json:"webhook_url"`
		EventTypes []string `json:"event_types"`
		IsActive   bool     `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}
	item, err := h.queries.UpdateTeamsWebhookConfig(r.Context(), db.UpdateTeamsWebhookConfigParams{
		ID: id, Name: body.Name, WebhookUrl: body.WebhookUrl, EventTypes: body.EventTypes, IsActive: body.IsActive,
	})
	if err != nil {
		http.Error(w, `{"error":"update failed"}`, http.StatusInternalServerError)
		return
	}
	auth.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) DeleteTeamsWebhook(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}
	_ = h.queries.DeleteTeamsWebhookConfig(r.Context(), id)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) SMTPStatus(w http.ResponseWriter, r *http.Request) {
	cfg := settings.ResolveSMTP(r.Context(), h.queries, h.cfg)
	auth.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"configured":      cfg.Configured(),
		"host":            cfg.Host,
		"port":            cfg.Port,
		"user":            cfg.User,
		"from":            cfg.From,
		"pass_configured": cfg.Pass != "",
	})
}

func sendSMTP(smtp settings.SMTPConfig, users []db.User, event Event) error {
	if !smtp.Configured() {
		return nil
	}
	m := gomail.NewMessage()
	m.SetHeader("From", smtp.From)
	var recipients []string
	for _, u := range users {
		recipients = append(recipients, u.Email)
	}
	if len(recipients) == 0 {
		return nil
	}
	m.SetHeader("To", recipients...)
	m.SetHeader("Subject", event.Title)
	m.SetBody("text/html", fmt.Sprintf("<h1>%s</h1><p>%s</p>", event.Title, event.Body))
	d := gomail.NewDialer(smtp.Host, smtp.Port, smtp.User, smtp.Pass)
	return d.DialAndSend(m)
}
