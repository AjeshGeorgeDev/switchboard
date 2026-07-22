package settings

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/switchboard/switchboard/internal/auth"
	"github.com/switchboard/switchboard/internal/config"
	"github.com/switchboard/switchboard/internal/db"
	"gopkg.in/gomail.v2"
)

const themePresetKey = "theme_preset"
const defaultTheme = "indigo-pulse"

var allowedThemes = map[string]struct{}{
	"indigo-pulse": {},
	"harbor-mist":  {},
	"signal-green": {},
	"ember-glow":   {},
	"midnight-sky": {},
	"violet-flare": {},
}

type Handler struct {
	queries *db.Queries
	cfg     config.Config
}

func NewHandler(queries *db.Queries, cfg config.Config) *Handler {
	return &Handler{queries: queries, cfg: cfg}
}

func (h *Handler) GetTheme(w http.ResponseWriter, r *http.Request) {
	preset := defaultTheme
	row, err := h.queries.GetAppSetting(r.Context(), themePresetKey)
	if err == nil && row.Value != "" {
		if _, ok := allowedThemes[row.Value]; ok {
			preset = row.Value
		}
	}
	auth.WriteJSON(w, http.StatusOK, map[string]string{"theme_preset": preset})
}

type updateThemeRequest struct {
	ThemePreset string `json:"theme_preset"`
}

func (h *Handler) UpdateTheme(w http.ResponseWriter, r *http.Request) {
	var req updateThemeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}
	if _, ok := allowedThemes[req.ThemePreset]; !ok {
		http.Error(w, `{"error":"unknown theme preset"}`, http.StatusBadRequest)
		return
	}
	row, err := h.queries.UpsertAppSetting(r.Context(), db.UpsertAppSettingParams{
		Key:   themePresetKey,
		Value: req.ThemePreset,
	})
	if err != nil {
		http.Error(w, `{"error":"failed to save theme"}`, http.StatusInternalServerError)
		return
	}
	auth.WriteJSON(w, http.StatusOK, map[string]string{"theme_preset": row.Value})
}

func (h *Handler) GetHarbor(w http.ResponseWriter, r *http.Request) {
	cfg := ResolveHarbor(r.Context(), h.queries, h.cfg)
	auth.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"url":                       cfg.URL,
		"user":                      cfg.User,
		"token_configured":          strings.TrimSpace(cfg.Token) != "",
		"webhook_secret_configured": strings.TrimSpace(cfg.WebhookSecret) != "",
		"api_configured":            cfg.APIConfigured(),
	})
}

type updateHarborRequest struct {
	URL                string `json:"url"`
	User               string `json:"user"`
	Token              string `json:"token"`
	WebhookSecret      string `json:"webhook_secret"`
	ClearToken         bool   `json:"clear_token"`
	ClearWebhookSecret bool   `json:"clear_webhook_secret"`
}

func (h *Handler) UpdateHarbor(w http.ResponseWriter, r *http.Request) {
	var req updateHarborRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	if err := upsertSetting(ctx, h.queries, KeyHarborURL, strings.TrimSpace(req.URL)); err != nil {
		http.Error(w, `{"error":"failed to save"}`, http.StatusInternalServerError)
		return
	}
	if err := upsertSetting(ctx, h.queries, KeyHarborUser, strings.TrimSpace(req.User)); err != nil {
		http.Error(w, `{"error":"failed to save"}`, http.StatusInternalServerError)
		return
	}
	if req.ClearToken {
		if err := clearSetting(ctx, h.queries, KeyHarborToken); err != nil {
			http.Error(w, `{"error":"failed to save"}`, http.StatusInternalServerError)
			return
		}
	} else if strings.TrimSpace(req.Token) != "" {
		if err := upsertSetting(ctx, h.queries, KeyHarborToken, strings.TrimSpace(req.Token)); err != nil {
			http.Error(w, `{"error":"failed to save"}`, http.StatusInternalServerError)
			return
		}
	}
	if req.ClearWebhookSecret {
		if err := clearSetting(ctx, h.queries, KeyHarborWebhookSecret); err != nil {
			http.Error(w, `{"error":"failed to save"}`, http.StatusInternalServerError)
			return
		}
	} else if strings.TrimSpace(req.WebhookSecret) != "" {
		if err := upsertSetting(ctx, h.queries, KeyHarborWebhookSecret, strings.TrimSpace(req.WebhookSecret)); err != nil {
			http.Error(w, `{"error":"failed to save"}`, http.StatusInternalServerError)
			return
		}
	}
	h.GetHarbor(w, r)
}

func (h *Handler) GetSMTP(w http.ResponseWriter, r *http.Request) {
	cfg := ResolveSMTP(r.Context(), h.queries, h.cfg)
	auth.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"configured":    cfg.Configured(),
		"host":          cfg.Host,
		"port":          cfg.Port,
		"user":          cfg.User,
		"from":          cfg.From,
		"pass_configured": strings.TrimSpace(cfg.Pass) != "",
	})
}

type updateSMTPRequest struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Pass     string `json:"pass"`
	From     string `json:"from"`
	ClearPass bool  `json:"clear_pass"`
}

func (h *Handler) UpdateSMTP(w http.ResponseWriter, r *http.Request) {
	var req updateSMTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	if err := upsertSetting(ctx, h.queries, KeySMTPHost, strings.TrimSpace(req.Host)); err != nil {
		http.Error(w, `{"error":"failed to save"}`, http.StatusInternalServerError)
		return
	}
	port := req.Port
	if port <= 0 {
		port = 587
	}
	if err := upsertSetting(ctx, h.queries, KeySMTPPort, strconv.Itoa(port)); err != nil {
		http.Error(w, `{"error":"failed to save"}`, http.StatusInternalServerError)
		return
	}
	if err := upsertSetting(ctx, h.queries, KeySMTPUser, strings.TrimSpace(req.User)); err != nil {
		http.Error(w, `{"error":"failed to save"}`, http.StatusInternalServerError)
		return
	}
	if err := upsertSetting(ctx, h.queries, KeySMTPFrom, strings.TrimSpace(req.From)); err != nil {
		http.Error(w, `{"error":"failed to save"}`, http.StatusInternalServerError)
		return
	}
	if req.ClearPass {
		if err := clearSetting(ctx, h.queries, KeySMTPPass); err != nil {
			http.Error(w, `{"error":"failed to save"}`, http.StatusInternalServerError)
			return
		}
	} else if strings.TrimSpace(req.Pass) != "" {
		if err := upsertSetting(ctx, h.queries, KeySMTPPass, strings.TrimSpace(req.Pass)); err != nil {
			http.Error(w, `{"error":"failed to save"}`, http.StatusInternalServerError)
			return
		}
	}
	h.GetSMTP(w, r)
}

type testHarborRequest struct {
	URL   string `json:"url"`
	User  string `json:"user"`
	Token string `json:"token"`
}

func (h *Handler) TestHarbor(w http.ResponseWriter, r *http.Request) {
	var req testHarborRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	cfg := ResolveHarbor(r.Context(), h.queries, h.cfg)
	if strings.TrimSpace(req.URL) != "" {
		cfg.URL = strings.TrimSpace(req.URL)
	}
	if strings.TrimSpace(req.User) != "" {
		cfg.User = strings.TrimSpace(req.User)
	}
	if strings.TrimSpace(req.Token) != "" {
		cfg.Token = strings.TrimSpace(req.Token)
	}

	if err := pingHarbor(r.Context(), cfg); err != nil {
		auth.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"ok":    false,
			"error": err.Error(),
		})
		return
	}
	auth.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"ok":      true,
		"message": "Connected to Harbor successfully",
	})
}

type testSMTPRequest struct {
	Host string `json:"host"`
	Port int    `json:"port"`
	User string `json:"user"`
	Pass string `json:"pass"`
	From string `json:"from"`
}

func (h *Handler) TestSMTP(w http.ResponseWriter, r *http.Request) {
	var req testSMTPRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	cfg := ResolveSMTP(r.Context(), h.queries, h.cfg)
	if strings.TrimSpace(req.Host) != "" {
		cfg.Host = strings.TrimSpace(req.Host)
	}
	if req.Port > 0 {
		cfg.Port = req.Port
	}
	if strings.TrimSpace(req.User) != "" {
		cfg.User = strings.TrimSpace(req.User)
	}
	if strings.TrimSpace(req.Pass) != "" {
		cfg.Pass = strings.TrimSpace(req.Pass)
	}
	if strings.TrimSpace(req.From) != "" {
		cfg.From = strings.TrimSpace(req.From)
	}
	if !cfg.Configured() {
		auth.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"ok":    false,
			"error": "SMTP host is required",
		})
		return
	}
	if strings.TrimSpace(cfg.From) == "" {
		auth.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"ok":    false,
			"error": "From address is required",
		})
		return
	}

	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}
	user, err := h.queries.GetUserByID(r.Context(), userID)
	if err != nil || strings.TrimSpace(user.Email) == "" {
		auth.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"ok":    false,
			"error": "Your account has no email address to send a test message to",
		})
		return
	}

	body := "This is a test email from Switchboard. SMTP is configured correctly."
	m := gomail.NewMessage()
	m.SetHeader("From", cfg.From)
	m.SetHeader("To", user.Email)
	m.SetHeader("Subject", "Switchboard SMTP test")
	m.SetBody("text/plain", body)
	d := gomail.NewDialer(cfg.Host, cfg.Port, cfg.User, cfg.Pass)
	sendErr := d.DialAndSend(m)

	status := "sent"
	errText := pgtype.Text{}
	recStatus := "sent"
	if sendErr != nil {
		status = "failed"
		recStatus = "failed"
		errText = pgtype.Text{String: sendErr.Error(), Valid: true}
	}
	logRow, logErr := h.queries.CreateEmailOutboundLog(r.Context(), db.CreateEmailOutboundLogParams{
		EventType:    "smtp_test",
		Subject:      "Switchboard SMTP test",
		BodyPreview:  body,
		Status:       status,
		ErrorMessage: errText,
		TriggeredBy:  pgtype.UUID{Bytes: userID, Valid: true},
	})
	if logErr == nil {
		_, _ = h.queries.CreateEmailOutboundRecipient(r.Context(), db.CreateEmailOutboundRecipientParams{
			LogID:        logRow.ID,
			Email:        user.Email,
			UserID:       pgtype.UUID{Bytes: user.ID, Valid: true},
			Status:       recStatus,
			ErrorMessage: errText,
		})
	}

	if sendErr != nil {
		auth.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"ok":    false,
			"error": sendErr.Error(),
		})
		return
	}
	auth.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"ok":      true,
		"message": "Test email sent to " + user.Email,
	})
}

type emailRecipientsResponse struct {
	WeeklyDigestRoles []string               `json:"weekly_digest_roles"`
	CriticalCVERoles  []string               `json:"critical_cve_roles"`
	Preview           []emailRecipientPreview `json:"preview"`
}

type emailRecipientPreview struct {
	ID     string   `json:"id"`
	Email  string   `json:"email"`
	Name   string   `json:"name"`
	Roles  []string `json:"roles"`
}

func (h *Handler) GetEmailRecipients(w http.ResponseWriter, r *http.Request) {
	digestRoles := EmailRecipientRoles(r.Context(), h.queries, "weekly_digest")
	criticalRoles := EmailRecipientRoles(r.Context(), h.queries, "critical_cve")
	roleSet := map[string]struct{}{}
	for _, rname := range digestRoles {
		roleSet[rname] = struct{}{}
	}
	for _, rname := range criticalRoles {
		roleSet[rname] = struct{}{}
	}
	allRoles := make([]string, 0, len(roleSet))
	for rname := range roleSet {
		allRoles = append(allRoles, rname)
	}
	users, err := h.queries.GetUsersByRoleNames(r.Context(), allRoles)
	if err != nil {
		http.Error(w, `{"error":"server error"}`, http.StatusInternalServerError)
		return
	}
	preview := make([]emailRecipientPreview, 0, len(users))
	for _, u := range users {
		roles, _ := h.queries.GetUserRoles(r.Context(), u.ID)
		names := make([]string, 0, len(roles))
		for _, role := range roles {
			names = append(names, role.Name)
		}
		name := u.Username
		if u.DisplayName.Valid && u.DisplayName.String != "" {
			name = u.DisplayName.String
		}
		preview = append(preview, emailRecipientPreview{
			ID:    u.ID.String(),
			Email: u.Email,
			Name:  name,
			Roles: names,
		})
	}
	auth.WriteJSON(w, http.StatusOK, emailRecipientsResponse{
		WeeklyDigestRoles: digestRoles,
		CriticalCVERoles:  criticalRoles,
		Preview:           preview,
	})
}

type updateEmailRecipientsRequest struct {
	WeeklyDigestRoles []string `json:"weekly_digest_roles"`
	CriticalCVERoles  []string `json:"critical_cve_roles"`
}

func (h *Handler) UpdateEmailRecipients(w http.ResponseWriter, r *http.Request) {
	var req updateEmailRecipientsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}
	if err := SaveEmailRecipientRoles(r.Context(), h.queries, req.WeeklyDigestRoles, req.CriticalCVERoles); err != nil {
		http.Error(w, `{"error":"failed to save"}`, http.StatusInternalServerError)
		return
	}
	h.GetEmailRecipients(w, r)
}

func (h *Handler) ListEmailLog(w http.ResponseWriter, r *http.Request) {
	eventType := r.URL.Query().Get("event_type")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	rows, err := h.queries.ListEmailOutboundLog(r.Context(), db.ListEmailOutboundLogParams{
		Column1: eventType,
		Limit:   int32(limit),
		Offset:  int32(offset),
	})
	if err != nil {
		http.Error(w, `{"error":"server error"}`, http.StatusInternalServerError)
		return
	}
	total, _ := h.queries.CountEmailOutboundLog(r.Context(), eventType)
	ids := make([]uuid.UUID, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}
	recByLog := map[uuid.UUID][]db.EmailOutboundRecipient{}
	if len(ids) > 0 {
		recs, err := h.queries.ListEmailOutboundRecipientsByLogIDs(r.Context(), ids)
		if err == nil {
			for _, rec := range recs {
				recByLog[rec.LogID] = append(recByLog[rec.LogID], rec)
			}
		}
	}
	items := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		recs := recByLog[row.ID]
		recipients := make([]map[string]interface{}, 0, len(recs))
		for _, rec := range recs {
			item := map[string]interface{}{
				"email":  rec.Email,
				"status": rec.Status,
			}
			if rec.ErrorMessage.Valid {
				item["error_message"] = rec.ErrorMessage.String
			}
			if rec.UserID.Valid {
				item["user_id"] = uuid.UUID(rec.UserID.Bytes).String()
			}
			recipients = append(recipients, item)
		}
		entry := map[string]interface{}{
			"id":              row.ID,
			"event_type":      row.EventType,
			"subject":         row.Subject,
			"body_preview":    row.BodyPreview,
			"status":          row.Status,
			"created_at":      row.CreatedAt,
			"recipient_count": len(recipients),
			"recipients":      recipients,
		}
		if row.ErrorMessage.Valid {
			entry["error_message"] = row.ErrorMessage.String
		}
		items = append(items, entry)
	}
	auth.WriteJSON(w, http.StatusOK, map[string]interface{}{"items": items, "total": total})
}
