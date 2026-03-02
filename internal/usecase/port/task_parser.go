package port

import (
	"context"

	"telegram-trello-bot/internal/domain/entity"
)

type TaskParser interface {
	Parse(ctx context.Context, rawMessage string) (*entity.Task, error)
}
