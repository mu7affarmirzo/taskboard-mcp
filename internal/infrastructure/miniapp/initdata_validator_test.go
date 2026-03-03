package miniapp

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"testing"
	"time"

	"telegram-trello-bot/internal/domain/domainerror"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testBotToken = "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"

func buildInitData(t *testing.T, botToken string, userID int64, username string, authDate int64) string {
	t.Helper()

	userData, err := json.Marshal(map[string]any{
		"id":         userID,
		"first_name": "Test",
		"username":   username,
	})
	require.NoError(t, err)

	values := url.Values{}
	values.Set("user", string(userData))
	values.Set("auth_date", fmt.Sprintf("%d", authDate))

	// Build data-check-string
	var pairs []string
	for key := range values {
		pairs = append(pairs, key+"="+values.Get(key))
	}
	sort.Strings(pairs)
	dataCheckString := strings.Join(pairs, "\n")

	// Compute HMAC
	secretKey := computeHMAC([]byte("WebAppData"), []byte(botToken))
	hashBytes := computeHMAC(secretKey, []byte(dataCheckString))
	values.Set("hash", hex.EncodeToString(hashBytes))

	return values.Encode()
}

func computeHMAC(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}

func TestInitDataValidator_HappyPath(t *testing.T) {
	now := time.Now()
	v := NewTelegramInitDataValidator(testBotToken)
	v.nowFunc = func() time.Time { return now }

	initData := buildInitData(t, testBotToken, 12345, "testuser", now.Unix())

	result, err := v.Validate(context.Background(), initData)

	assert.NoError(t, err)
	assert.Equal(t, int64(12345), result.TelegramID)
	assert.Equal(t, "Test", result.FirstName)
	assert.Equal(t, "testuser", result.Username)
	assert.Equal(t, now.Unix(), result.AuthDate)
}

func TestInitDataValidator_InvalidHash(t *testing.T) {
	now := time.Now()
	v := NewTelegramInitDataValidator(testBotToken)
	v.nowFunc = func() time.Time { return now }

	initData := buildInitData(t, "wrong-token", 12345, "testuser", now.Unix())

	result, err := v.Validate(context.Background(), initData)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, domainerror.ErrInvalidInitData)
	assert.Contains(t, err.Error(), "hash mismatch")
}

func TestInitDataValidator_MissingHash(t *testing.T) {
	v := NewTelegramInitDataValidator(testBotToken)

	result, err := v.Validate(context.Background(), "auth_date=123456&user=%7B%7D")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, domainerror.ErrInvalidInitData)
	assert.Contains(t, err.Error(), "missing hash")
}

func TestInitDataValidator_AuthDateTooOld(t *testing.T) {
	now := time.Now()
	v := NewTelegramInitDataValidator(testBotToken)
	v.nowFunc = func() time.Time { return now }

	oldDate := now.Add(-25 * time.Hour).Unix()
	initData := buildInitData(t, testBotToken, 12345, "testuser", oldDate)

	result, err := v.Validate(context.Background(), initData)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, domainerror.ErrInvalidInitData)
	assert.Contains(t, err.Error(), "auth_date too old")
}

func TestInitDataValidator_MissingUser(t *testing.T) {
	now := time.Now()
	v := NewTelegramInitDataValidator(testBotToken)
	v.nowFunc = func() time.Time { return now }

	values := url.Values{}
	values.Set("auth_date", fmt.Sprintf("%d", now.Unix()))

	var pairs []string
	for key := range values {
		pairs = append(pairs, key+"="+values.Get(key))
	}
	sort.Strings(pairs)
	dataCheckString := strings.Join(pairs, "\n")

	secretKey := computeHMAC([]byte("WebAppData"), []byte(testBotToken))
	hashBytes := computeHMAC(secretKey, []byte(dataCheckString))
	values.Set("hash", hex.EncodeToString(hashBytes))

	result, err := v.Validate(context.Background(), values.Encode())

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, domainerror.ErrInvalidInitData)
	assert.Contains(t, err.Error(), "missing user")
}

func TestInitDataValidator_MalformedQueryString(t *testing.T) {
	v := NewTelegramInitDataValidator(testBotToken)

	result, err := v.Validate(context.Background(), "%%invalid")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, domainerror.ErrInvalidInitData)
}
