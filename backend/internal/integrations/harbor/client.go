package harbor

import (
	"context"

	"github.com/switchboard/switchboard/internal/config"
)

type Client struct {
	cfg config.Config
}

func NewClient(cfg config.Config) *Client {
	return &Client{cfg: cfg}
}

func (c *Client) ListRepositories(ctx context.Context) ([]string, error) {
	if c.cfg.HarborURL == "" {
		return []string{}, nil
	}
	// Placeholder: integrate Harbor REST API when credentials are available
	_ = ctx
	return []string{}, nil
}
