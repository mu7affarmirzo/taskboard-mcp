package usecase

import (
	"context"
	"fmt"

	"telegram-trello-bot/internal/domain/domainerror"
	"telegram-trello-bot/internal/domain/valueobject"
	"telegram-trello-bot/internal/usecase/dto"
	"telegram-trello-bot/internal/usecase/port"
)

type GetCardUseCase struct {
	cardManager port.CardManager
	userRepo    port.UserRepository
}

func NewGetCardUseCase(cardManager port.CardManager, userRepo port.UserRepository) *GetCardUseCase {
	return &GetCardUseCase{cardManager: cardManager, userRepo: userRepo}
}

func (uc *GetCardUseCase) Execute(ctx context.Context, telegramID int64, cardID string) (*dto.GetCardOutput, error) {
	user, err := uc.userRepo.FindByTelegramID(ctx, valueobject.TelegramID(telegramID))
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}
	if !user.HasTrelloToken() {
		return nil, domainerror.ErrTrelloNotConnected
	}

	card, err := uc.cardManager.GetCard(ctx, user.TrelloToken(), cardID)
	if err != nil {
		return nil, fmt.Errorf("get card: %w", err)
	}

	return &dto.GetCardOutput{
		ID:          card.ID,
		Title:       card.Title,
		Description: card.Description,
		URL:         card.URL,
		ListID:      card.ListID,
		Due:         card.Due,
		Labels:      card.Labels,
		Members:     card.Members,
	}, nil
}
