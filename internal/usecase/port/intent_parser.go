package port

import (
	"context"

	"telegram-trello-bot/internal/usecase/dto"
)

type IntentParser interface {
	ParseIntent(ctx context.Context, rawMessage string) (*dto.IntentOutput, error)
}
