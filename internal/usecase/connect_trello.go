package usecase

import (
	"context"
	"fmt"

	"telegram-trello-bot/internal/domain/domainerror"
	"telegram-trello-bot/internal/domain/valueobject"
	"telegram-trello-bot/internal/usecase/dto"
	"telegram-trello-bot/internal/usecase/port"
)

type ConnectTrelloUseCase struct {
	userRepo port.UserRepository
}

func NewConnectTrelloUseCase(userRepo port.UserRepository) *ConnectTrelloUseCase {
	return &ConnectTrelloUseCase{userRepo: userRepo}
}

func (uc *ConnectTrelloUseCase) Execute(ctx context.Context, input dto.ConnectTrelloInput) (*dto.ConnectTrelloOutput, error) {
	if input.Token == "" {
		return nil, domainerror.ErrEmptyTrelloToken
	}

	user, err := uc.userRepo.FindByTelegramID(ctx, valueobject.TelegramID(input.TelegramID))
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}

	user.SetTrelloToken(input.Token)
	if err := uc.userRepo.Save(ctx, user); err != nil {
		return nil, fmt.Errorf("save user: %w", err)
	}

	return &dto.ConnectTrelloOutput{Connected: true}, nil
}
