package health

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "modernc.org/sqlite"
)

func TestHealthHandler_Healthy(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	handler := NewHealthHandler(db)
	req := httptest.NewRequest("GET", "/healthz", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var resp response
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if resp.Status != "ok" {
		t.Fatalf("expected status 'ok', got '%s'", resp.Status)
	}
}

func TestHealthHandler_Unhealthy(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	// Close the DB so Ping fails
	_ = db.Close()

	handler := NewHealthHandler(db)
	req := httptest.NewRequest("GET", "/healthz", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rr.Code)
	}

	var resp response
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if resp.Status != "unhealthy" {
		t.Fatalf("expected status 'unhealthy', got '%s'", resp.Status)
	}
}

func TestHealthHandler_ContentType(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	handler := NewHealthHandler(db)
	req := httptest.NewRequest("GET", "/healthz", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	ct := rr.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Fatalf("expected content-type 'application/json', got '%s'", ct)
	}
}
