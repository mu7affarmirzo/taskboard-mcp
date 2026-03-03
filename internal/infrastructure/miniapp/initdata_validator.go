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
	"strconv"
	"strings"
	"time"

	"telegram-trello-bot/internal/domain/domainerror"
	"telegram-trello-bot/internal/usecase/port"
)

type TelegramInitDataValidator struct {
	botToken  string
	maxAge    time.Duration
	nowFunc   func() time.Time
}

func NewTelegramInitDataValidator(botToken string) *TelegramInitDataValidator {
	return &TelegramInitDataValidator{
		botToken: botToken,
		maxAge:   24 * time.Hour,
		nowFunc:  time.Now,
	}
}

func (v *TelegramInitDataValidator) Validate(_ context.Context, initData string) (*port.InitDataResult, error) {
	values, err := url.ParseQuery(initData)
	if err != nil {
		return nil, fmt.Errorf("%w: malformed query string", domainerror.ErrInvalidInitData)
	}

	hash := values.Get("hash")
	if hash == "" {
		return nil, fmt.Errorf("%w: missing hash", domainerror.ErrInvalidInitData)
	}

	// Build data-check-string: sorted key=value pairs excluding "hash"
	var pairs []string
	for key := range values {
		if key == "hash" {
			continue
		}
		pairs = append(pairs, key+"="+values.Get(key))
	}
	sort.Strings(pairs)
	dataCheckString := strings.Join(pairs, "\n")

	// Compute HMAC-SHA256
	secretKey := hmacSHA256([]byte("WebAppData"), []byte(v.botToken))
	computedHash := hmacSHA256(secretKey, []byte(dataCheckString))
	computedHashHex := hex.EncodeToString(computedHash)

	if !hmac.Equal([]byte(computedHashHex), []byte(hash)) {
		return nil, fmt.Errorf("%w: hash mismatch", domainerror.ErrInvalidInitData)
	}

	// Parse auth_date
	authDateStr := values.Get("auth_date")
	if authDateStr == "" {
		return nil, fmt.Errorf("%w: missing auth_date", domainerror.ErrInvalidInitData)
	}
	authDate, err := strconv.ParseInt(authDateStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid auth_date", domainerror.ErrInvalidInitData)
	}

	// Reject if older than maxAge
	authTime := time.Unix(authDate, 0)
	if v.nowFunc().Sub(authTime) > v.maxAge {
		return nil, fmt.Errorf("%w: auth_date too old", domainerror.ErrInvalidInitData)
	}

	// Parse user JSON
	userJSON := values.Get("user")
	if userJSON == "" {
		return nil, fmt.Errorf("%w: missing user", domainerror.ErrInvalidInitData)
	}

	var userData struct {
		ID        int64  `json:"id"`
		FirstName string `json:"first_name"`
		Username  string `json:"username"`
	}
	if err := json.Unmarshal([]byte(userJSON), &userData); err != nil {
		return nil, fmt.Errorf("%w: invalid user JSON", domainerror.ErrInvalidInitData)
	}

	return &port.InitDataResult{
		TelegramID: userData.ID,
		FirstName:  userData.FirstName,
		Username:   userData.Username,
		AuthDate:   authDate,
	}, nil
}

func hmacSHA256(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}
