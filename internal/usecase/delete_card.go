package usecase

import (
	"context"
	"fmt"

	"telegram-trello-bot/internal/domain/domainerror"
	"telegram-trello-bot/internal/domain/valueobject"
	"telegram-trello-bot/internal/usecase/port"
)

type DeleteCardUseCase struct {
	cardManager port.CardManager
	userRepo    port.UserRepository
}

func NewDeleteCardUseCase(cardManager port.CardManager, userRepo port.UserRepository) *DeleteCardUseCase {
	return &DeleteCardUseCase{cardManager: cardManager, userRepo: userRepo}
}

func (uc *DeleteCardUseCase) Execute(ctx context.Context, telegramID int64, cardID string) error {
	user, err := uc.userRepo.FindByTelegramID(ctx, valueobject.TelegramID(telegramID))
	if err != nil {
		return fmt.Errorf("find user: %w", err)
	}
	if !user.HasTrelloToken() {
		return domainerror.ErrTrelloNotConnected
	}

	if err := uc.cardManager.DeleteCard(ctx, user.TrelloToken(), cardID); err != nil {
		return fmt.Errorf("delete card: %w", err)
	}
	return nil
}
