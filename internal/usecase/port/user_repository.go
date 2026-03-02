package port

import (
	"context"

	"telegram-trello-bot/internal/domain/entity"
	"telegram-trello-bot/internal/domain/valueobject"
)

type UserRepository interface {
	FindByTelegramID(ctx context.Context, id valueobject.TelegramID) (*entity.User, error)
	Save(ctx context.Context, user *entity.User) error
}
