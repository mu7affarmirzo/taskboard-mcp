package port

import "context"

type SessionClaims struct {
	TelegramID int64
	Username   string
}

type SessionManager interface {
	CreateToken(ctx context.Context, claims SessionClaims) (string, error)
	ValidateToken(ctx context.Context, token string) (*SessionClaims, error)
}
