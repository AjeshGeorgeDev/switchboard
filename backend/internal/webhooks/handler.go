package webhooks

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/switchboard/switchboard/internal/auth"
	"github.com/switchboard/switchboard/internal/config"
	"github.com/switchboard/switchboard/internal/db"
	"github.com/switchboard/switchboard/internal/jobs"
	"github.com/switchboard/switchboard/internal/settings"
)

type Handler struct {
	client  *asynq.Client
	cfg     config.Config
	queries *db.Queries
}

func NewHandler(client *asynq.Client, cfg config.Config, queries *db.Queries) *Handler {
	return &Handler{client: client, cfg: cfg, queries: queries}
}

type taskEnvelope struct {
	EventID string          `json:"event_id"`
	Body    json.RawMessage `json:"body"`
}

func (h *Handler) Endpoints(w http.ResponseWriter, r *http.Request) {
	base := strings.TrimRight(h.cfg.AppBaseURL, "/")
	harborCfg := settings.ResolveHarbor(r.Context(), h.queries, h.cfg)
	auth.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"harbor_url":               base + "/webhooks/harbor",
		"trivy_url":                base + "/webhooks/trivy",
		"harbor_secret_configured": harborCfg.WebhookSecret != "",
		"trivy_secret_configured":  h.cfg.TrivyWebhookSecret != "",
		"harbor_api_configured":    harborCfg.APIConfigured(),
		"cve_pull_enabled":         h.cfg.CVEPullEnabled,
		"cve_pull_configured":      h.cfg.TrivyURL != "" && h.cfg.TrivyToken != "",
		"cve_pull_cron":            h.cfg.CVEPullCron,
	})
}

func (h *Handler) ListEvents(w http.ResponseWriter, r *http.Request) {
	limit, offset := pagination(r)
	source := r.URL.Query().Get("source")
	events, err := h.queries.ListWebhookEvents(r.Context(), db.ListWebhookEventsParams{
		Column1: source,
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		http.Error(w, `{"error":"server error"}`, http.StatusInternalServerError)
		return
	}
	count, _ := h.queries.CountWebhookEvents(r.Context(), source)
	auth.WriteJSON(w, http.StatusOK, map[string]interface{}{"items": events, "total": count})
}

func (h *Handler) GetEvent(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}
	event, err := h.queries.GetWebhookEventByID(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	auth.WriteJSON(w, http.StatusOK, event)
}

func (h *Handler) Harbor(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, `{"error":"read failed"}`, http.StatusBadRequest)
		return
	}
	harborCfg := settings.ResolveHarbor(r.Context(), h.queries, h.cfg)
	if !verifySecret(body, r.Header.Get("X-Webhook-Signature"), harborCfg.WebhookSecret) {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}
	h.receive(r.Context(), w, body, db.WebhookSourceHarbor, jobs.TypeProcessHarborWebhook)
}

func (h *Handler) Trivy(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, `{"error":"read failed"}`, http.StatusBadRequest)
		return
	}
	if !verifySecret(body, r.Header.Get("X-Webhook-Signature"), h.cfg.TrivyWebhookSecret) {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}
	h.receive(r.Context(), w, body, db.WebhookSourceTrivy, jobs.TypeProcessTrivyWebhook)
}

func (h *Handler) receive(ctx context.Context, w http.ResponseWriter, body []byte, source db.WebhookSource, taskType string) {
	preview := string(body)
	if len(preview) > 500 {
		preview = preview[:500] + "…"
	}
	event, err := h.queries.CreateWebhookEvent(ctx, db.CreateWebhookEventParams{
		Source:         source,
		Payload:        body,
		PayloadPreview: pgtype.Text{String: preview, Valid: true},
	})
	if err != nil {
		http.Error(w, `{"error":"storage failed"}`, http.StatusInternalServerError)
		return
	}
	envelope, _ := json.Marshal(taskEnvelope{EventID: event.ID.String(), Body: body})
	task := asynq.NewTask(taskType, envelope)
	if _, err := h.client.Enqueue(task); err != nil {
		_ = h.queries.UpdateWebhookEventStatus(ctx, db.UpdateWebhookEventStatusParams{
			ID:           event.ID,
			Status:       db.WebhookEventStatusFailed,
			ErrorMessage: pgtype.Text{String: "enqueue failed: " + err.Error(), Valid: true},
		})
		http.Error(w, `{"error":"enqueue failed"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "accepted", "event_id": event.ID.String()})
}

func verifySecret(body []byte, sig, secret string) bool {
	if secret == "" {
		return true
	}
	if sig == "" {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(sig), []byte(expected))
}

func pagination(r *http.Request) (int32, int32) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	return int32(limit), int32(offset)
}
