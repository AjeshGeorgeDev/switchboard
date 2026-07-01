package security

import (
	"net/http/httptest"
	"testing"
)

func TestPaginationDefaults(t *testing.T) {
	req := httptest.NewRequest("GET", "/?offset=5", nil)
	limit, offset := pagination(req)
	if limit != 50 || offset != 5 {
		t.Fatalf("got limit=%d offset=%d", limit, offset)
	}
}

func TestPaginationInvalidLimitUsesDefault(t *testing.T) {
	req := httptest.NewRequest("GET", "/?limit=500", nil)
	limit, _ := pagination(req)
	if limit != 50 {
		t.Fatalf("expected default 50 for out-of-range limit, got %d", limit)
	}
}

func TestPaginationValidLimit(t *testing.T) {
	req := httptest.NewRequest("GET", "/?limit=100", nil)
	limit, _ := pagination(req)
	if limit != 100 {
		t.Fatalf("expected 100, got %d", limit)
	}
}
