package usecase

import (
	"context"

	"telegram-trello-bot/internal/domain/valueobject"
	"telegram-trello-bot/internal/usecase/port"
)

type SelectBoardUseCase struct {
	userRepo port.UserRepository
}

func NewSelectBoardUseCase(userRepo port.UserRepository) *SelectBoardUseCase {
	return &SelectBoardUseCase{userRepo: userRepo}
}

func (uc *SelectBoardUseCase) Execute(ctx context.Context, telegramID int64, boardID string) error {
	user, err := uc.userRepo.FindByTelegramID(ctx, valueobject.TelegramID(telegramID))
	if err != nil {
		return err
	}
	user.SetDefaultBoard(boardID)
	return uc.userRepo.Save(ctx, user)
}
