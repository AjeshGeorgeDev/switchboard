package audit

import (
	"net/http/httptest"
	"testing"
)

func TestPaginationDefaults(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	limit, offset := pagination(req)
	if limit != 50 || offset != 0 {
		t.Fatalf("got limit=%d offset=%d", limit, offset)
	}
}

func TestPaginationCustom(t *testing.T) {
	req := httptest.NewRequest("GET", "/?limit=10&offset=20", nil)
	limit, offset := pagination(req)
	if limit != 10 || offset != 20 {
		t.Fatalf("got limit=%d offset=%d", limit, offset)
	}
}
