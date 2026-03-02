package usecase

import (
	"context"

	"telegram-trello-bot/internal/domain/valueobject"
	"telegram-trello-bot/internal/usecase/port"
)

type SelectListUseCase struct {
	userRepo port.UserRepository
}

func NewSelectListUseCase(userRepo port.UserRepository) *SelectListUseCase {
	return &SelectListUseCase{userRepo: userRepo}
}

func (uc *SelectListUseCase) Execute(ctx context.Context, telegramID int64, listID string) error {
	user, err := uc.userRepo.FindByTelegramID(ctx, valueobject.TelegramID(telegramID))
	if err != nil {
		return err
	}
	user.SetDefaultList(listID)
	return uc.userRepo.Save(ctx, user)
}
