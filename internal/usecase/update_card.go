package usecase

import (
	"context"
	"fmt"

	"telegram-trello-bot/internal/domain/domainerror"
	"telegram-trello-bot/internal/domain/valueobject"
	"telegram-trello-bot/internal/usecase/dto"
	"telegram-trello-bot/internal/usecase/port"
)

type UpdateCardUseCase struct {
	cardManager port.CardManager
	userRepo    port.UserRepository
}

func NewUpdateCardUseCase(cardManager port.CardManager, userRepo port.UserRepository) *UpdateCardUseCase {
	return &UpdateCardUseCase{cardManager: cardManager, userRepo: userRepo}
}

func (uc *UpdateCardUseCase) Execute(ctx context.Context, telegramID int64, cardID string, input dto.UpdateCardInput) error {
	user, err := uc.userRepo.FindByTelegramID(ctx, valueobject.TelegramID(telegramID))
	if err != nil {
		return fmt.Errorf("find user: %w", err)
	}
	if !user.HasTrelloToken() {
		return domainerror.ErrTrelloNotConnected
	}

	params := port.UpdateCardParams{
		Name:      input.Title,
		Desc:      input.Description,
		IDList:    input.ListID,
		Due:       input.Due,
		IDLabels:  input.LabelIDs,
		IDMembers: input.MemberIDs,
	}

	if err := uc.cardManager.UpdateCard(ctx, user.TrelloToken(), cardID, params); err != nil {
		return fmt.Errorf("update card: %w", err)
	}
	return nil
}
