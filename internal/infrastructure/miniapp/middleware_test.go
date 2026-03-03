package miniapp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"telegram-trello-bot/internal/usecase/port"

	"github.com/stretchr/testify/assert"
)

func TestCORSMiddleware_Preflight(t *testing.T) {
	handler := CORSMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called on OPTIONS")
	}))

	req := httptest.NewRequest(http.MethodOptions, "/api/boards", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Equal(t, "*", rec.Header().Get("Access-Control-Allow-Origin"))
	assert.Contains(t, rec.Header().Get("Access-Control-Allow-Headers"), "Authorization")
}

func TestCORSMiddleware_PassesThrough(t *testing.T) {
	called := false
	handler := CORSMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/boards", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.True(t, called)
	assert.Equal(t, "*", rec.Header().Get("Access-Control-Allow-Origin"))
}

func TestAuthMiddleware_SkipsAuthEndpoint(t *testing.T) {
	sessionMgr := NewJWTSessionManager("test-secret")
	called := false
	handler := AuthMiddleware(sessionMgr)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/auth", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.True(t, called)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAuthMiddleware_MissingHeader(t *testing.T) {
	sessionMgr := NewJWTSessionManager("test-secret")
	handler := AuthMiddleware(sessionMgr)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/boards", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthMiddleware_InvalidFormat(t *testing.T) {
	sessionMgr := NewJWTSessionManager("test-secret")
	handler := AuthMiddleware(sessionMgr)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/boards", nil)
	req.Header.Set("Authorization", "Basic abc123")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	sessionMgr := NewJWTSessionManager("test-secret")

	token, err := sessionMgr.CreateToken(context.Background(), port.SessionClaims{
		TelegramID: 12345,
		Username:   "testuser",
	})
	assert.NoError(t, err)

	var capturedID int64
	handler := AuthMiddleware(sessionMgr)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedID = TelegramIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/boards", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, int64(12345), capturedID)
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	sessionMgr := NewJWTSessionManager("test-secret")
	handler := AuthMiddleware(sessionMgr)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/boards", nil)
	req.Header.Set("Authorization", "Bearer invalid-jwt-token")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestJSONMiddleware_SetsContentType(t *testing.T) {
	handler := JSONMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/boards", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
}

func TestTelegramIDFromContext_NoValue(t *testing.T) {
	ctx := context.Background()
	id := TelegramIDFromContext(ctx)
	assert.Equal(t, int64(0), id)
}
