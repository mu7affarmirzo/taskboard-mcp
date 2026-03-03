package usecase

import (
	"context"
	"fmt"

	"telegram-trello-bot/internal/domain/domainerror"
	"telegram-trello-bot/internal/domain/valueobject"
	"telegram-trello-bot/internal/usecase/dto"
	"telegram-trello-bot/internal/usecase/port"
)

type ParseIntentUseCase struct {
	parser   port.IntentParser
	userRepo port.UserRepository
}

func NewParseIntentUseCase(
	parser port.IntentParser,
	userRepo port.UserRepository,
) *ParseIntentUseCase {
	return &ParseIntentUseCase{
		parser:   parser,
		userRepo: userRepo,
	}
}

func (uc *ParseIntentUseCase) Execute(
	ctx context.Context,
	input dto.IntentInput,
) (*dto.IntentOutput, error) {
	user, err := uc.userRepo.FindByTelegramID(ctx, valueobject.TelegramID(input.TelegramID))
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}
	if !user.HasTrelloToken() {
		return nil, domainerror.ErrTrelloNotConnected
	}
	if !user.HasBoardConfigured() {
		return nil, domainerror.ErrBoardNotSet
	}
	if !user.HasListConfigured() {
		return nil, domainerror.ErrListNotSet
	}

	intent, err := uc.parser.ParseIntent(ctx, input.RawMessage)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domainerror.ErrParsingFailed, err)
	}

	if _, err := valueobject.NewAction(intent.Action); err != nil {
		return nil, fmt.Errorf("%w: %s", domainerror.ErrUnknownAction, intent.Action)
	}

	return intent, nil
}
