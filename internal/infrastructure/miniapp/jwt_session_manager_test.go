package miniapp

import (
	"context"
	"testing"

	"telegram-trello-bot/internal/domain/domainerror"
	"telegram-trello-bot/internal/usecase/port"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTSessionManager_CreateAndValidate(t *testing.T) {
	mgr := NewJWTSessionManager("test-secret-key")

	token, err := mgr.CreateToken(context.Background(), port.SessionClaims{
		TelegramID: 12345,
		Username:   "testuser",
	})

	require.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := mgr.ValidateToken(context.Background(), token)

	assert.NoError(t, err)
	assert.Equal(t, int64(12345), claims.TelegramID)
	assert.Equal(t, "testuser", claims.Username)
}

func TestJWTSessionManager_InvalidToken(t *testing.T) {
	mgr := NewJWTSessionManager("test-secret-key")

	claims, err := mgr.ValidateToken(context.Background(), "invalid-token")

	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.ErrorIs(t, err, domainerror.ErrSessionExpired)
}

func TestJWTSessionManager_WrongSecret(t *testing.T) {
	mgr1 := NewJWTSessionManager("secret-1")
	mgr2 := NewJWTSessionManager("secret-2")

	token, err := mgr1.CreateToken(context.Background(), port.SessionClaims{
		TelegramID: 12345,
		Username:   "testuser",
	})
	require.NoError(t, err)

	claims, err := mgr2.ValidateToken(context.Background(), token)

	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.ErrorIs(t, err, domainerror.ErrSessionExpired)
}

func TestJWTSessionManager_EmptyUsername(t *testing.T) {
	mgr := NewJWTSessionManager("test-secret")

	token, err := mgr.CreateToken(context.Background(), port.SessionClaims{
		TelegramID: 99999,
		Username:   "",
	})
	require.NoError(t, err)

	claims, err := mgr.ValidateToken(context.Background(), token)

	assert.NoError(t, err)
	assert.Equal(t, int64(99999), claims.TelegramID)
	assert.Equal(t, "", claims.Username)
}
