package settings

import (
	"encoding/json"
	"net/http"

	"github.com/switchboard/switchboard/internal/auth"
	"github.com/switchboard/switchboard/internal/db"
)

const themePresetKey = "theme_preset"
const defaultTheme = "indigo-pulse"

var allowedThemes = map[string]struct{}{
	"indigo-pulse":  {},
	"harbor-mist":   {},
	"signal-green":  {},
	"ember-glow":    {},
	"midnight-sky":  {},
	"violet-flare":  {},
}

type Handler struct {
	queries *db.Queries
}

func NewHandler(queries *db.Queries) *Handler {
	return &Handler{queries: queries}
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
