package settings

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

func pingHarbor(ctx context.Context, cfg HarborConfig) error {
	if !cfg.APIConfigured() {
		return fmt.Errorf("Harbor URL and credentials are required")
	}
	user := strings.TrimSpace(cfg.User)
	token := strings.TrimSpace(cfg.Token)
	var basic string
	switch {
	case user != "":
		basic = user + ":" + token
	case strings.Contains(token, ":"):
		basic = token
	default:
		return fmt.Errorf("set Harbor user to the robot name and token to the secret only")
	}

	base := strings.TrimRight(strings.TrimSpace(cfg.URL), "/")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+"/api/v2.0/projects?page_size=1", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(basic)))

	client := &http.Client{Timeout: 20 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("unreachable: %w", err)
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	if res.StatusCode == http.StatusUnauthorized || res.StatusCode == http.StatusForbidden {
		return fmt.Errorf("%s — check robot username and token", res.Status)
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		msg := string(body)
		if len(msg) > 160 {
			msg = msg[:160] + "…"
		}
		return fmt.Errorf("harbor API %s: %s", res.Status, msg)
	}
	return nil
}
