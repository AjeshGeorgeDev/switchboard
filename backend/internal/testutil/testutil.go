package testutil

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/switchboard/switchboard/internal/db"
)

func DatabaseURL() string {
	if u := os.Getenv("TEST_DATABASE_URL"); u != "" {
		return u
	}
	return os.Getenv("DATABASE_URL")
}

func OpenPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := DatabaseURL()
	if url == "" {
		t.Skip("TEST_DATABASE_URL or DATABASE_URL not set; skipping integration test")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatalf("connect db: %v", err)
	}
	t.Cleanup(pool.Close)
	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("ping db: %v", err)
	}
	return pool
}

func Queries(t *testing.T) *db.Queries {
	t.Helper()
	return db.New(OpenPool(t))
}

func QueriesAndPool(t *testing.T) (*db.Queries, *pgxpool.Pool) {
	t.Helper()
	pool := OpenPool(t)
	return db.New(pool), pool
}

func DecodeJSON[T any](t *testing.T, resp *http.Response) T {
	t.Helper()
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	var out T
	if err := json.Unmarshal(body, &out); err != nil {
		t.Fatalf("decode json: %v\nbody: %s", err, body)
	}
	return out
}

func NewRequest(method, target string, body io.Reader) *http.Request {
	req, _ := http.NewRequest(method, target, body)
	return req
}
