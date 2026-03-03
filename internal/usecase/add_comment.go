package usecase

import (
	"context"
	"fmt"

	"telegram-trello-bot/internal/domain/domainerror"
	"telegram-trello-bot/internal/domain/valueobject"
	"telegram-trello-bot/internal/usecase/port"
)

type AddCommentUseCase struct {
	cardManager port.CardManager
	userRepo    port.UserRepository
}

func NewAddCommentUseCase(cardManager port.CardManager, userRepo port.UserRepository) *AddCommentUseCase {
	return &AddCommentUseCase{cardManager: cardManager, userRepo: userRepo}
}

func (uc *AddCommentUseCase) Execute(ctx context.Context, telegramID int64, cardID string, text string) error {
	user, err := uc.userRepo.FindByTelegramID(ctx, valueobject.TelegramID(telegramID))
	if err != nil {
		return fmt.Errorf("find user: %w", err)
	}
	if !user.HasTrelloToken() {
		return domainerror.ErrTrelloNotConnected
	}

	if err := uc.cardManager.AddComment(ctx, user.TrelloToken(), cardID, text); err != nil {
		return fmt.Errorf("add comment: %w", err)
	}
	return nil
}
