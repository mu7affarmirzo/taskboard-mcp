package miniapp

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"telegram-trello-bot/internal/domain/domainerror"
	"telegram-trello-bot/internal/usecase/port"
)

type JWTSessionManager struct {
	secret []byte
	expiry time.Duration
}

func NewJWTSessionManager(secret string) *JWTSessionManager {
	return &JWTSessionManager{
		secret: []byte(secret),
		expiry: 24 * time.Hour,
	}
}

type customClaims struct {
	TelegramID int64  `json:"telegram_id"`
	Username   string `json:"username"`
	jwt.RegisteredClaims
}

func (m *JWTSessionManager) CreateToken(_ context.Context, claims port.SessionClaims) (string, error) {
	now := time.Now()
	c := customClaims{
		TelegramID: claims.TelegramID,
		Username:   claims.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.expiry)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return signed, nil
}

func (m *JWTSessionManager) ValidateToken(_ context.Context, tokenStr string) (*port.SessionClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &customClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domainerror.ErrSessionExpired, err)
	}

	claims, ok := token.Claims.(*customClaims)
	if !ok || !token.Valid {
		return nil, domainerror.ErrSessionExpired
	}

	return &port.SessionClaims{
		TelegramID: claims.TelegramID,
		Username:   claims.Username,
	}, nil
}
