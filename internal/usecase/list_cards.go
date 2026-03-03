package usecase

import (
	"context"
	"fmt"

	"telegram-trello-bot/internal/domain/domainerror"
	"telegram-trello-bot/internal/domain/valueobject"
	"telegram-trello-bot/internal/usecase/dto"
	"telegram-trello-bot/internal/usecase/port"
)

type ListCardsUseCase struct {
	board    port.TaskBoard
	userRepo port.UserRepository
}

func NewListCardsUseCase(board port.TaskBoard, userRepo port.UserRepository) *ListCardsUseCase {
	return &ListCardsUseCase{board: board, userRepo: userRepo}
}

func (uc *ListCardsUseCase) Execute(ctx context.Context, telegramID int64, listID string) (*dto.ListCardsOutput, error) {
	user, err := uc.userRepo.FindByTelegramID(ctx, valueobject.TelegramID(telegramID))
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}
	if !user.HasTrelloToken() {
		return nil, domainerror.ErrTrelloNotConnected
	}

	cards, err := uc.board.GetCards(ctx, user.TrelloToken(), listID)
	if err != nil {
		return nil, fmt.Errorf("get cards: %w", err)
	}

	items := make([]dto.CardItem, len(cards))
	for i, c := range cards {
		items[i] = dto.CardItem{ID: c.CardID, Title: c.Title, URL: c.CardURL, ListID: c.ListID}
	}
	return &dto.ListCardsOutput{Cards: items}, nil
}
