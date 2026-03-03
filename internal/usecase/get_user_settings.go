package usecase

import (
	"context"
	"fmt"

	"telegram-trello-bot/internal/domain/valueobject"
	"telegram-trello-bot/internal/usecase/dto"
	"telegram-trello-bot/internal/usecase/port"
)

type GetUserSettingsUseCase struct {
	userRepo port.UserRepository
}

func NewGetUserSettingsUseCase(userRepo port.UserRepository) *GetUserSettingsUseCase {
	return &GetUserSettingsUseCase{userRepo: userRepo}
}

func (uc *GetUserSettingsUseCase) Execute(ctx context.Context, telegramID int64) (*dto.GetUserSettingsOutput, error) {
	user, err := uc.userRepo.FindByTelegramID(ctx, valueobject.TelegramID(telegramID))
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}

	return &dto.GetUserSettingsOutput{
		TrelloConnected: user.HasTrelloToken(),
		DefaultBoardID:  user.DefaultBoard(),
		DefaultListID:   user.DefaultList(),
	}, nil
}
