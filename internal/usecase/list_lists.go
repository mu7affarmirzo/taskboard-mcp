package usecase

import (
	"context"
	"fmt"

	"telegram-trello-bot/internal/domain/domainerror"
	"telegram-trello-bot/internal/domain/valueobject"
	"telegram-trello-bot/internal/usecase/dto"
	"telegram-trello-bot/internal/usecase/port"
)

type ListListsUseCase struct {
	board    port.TaskBoard
	userRepo port.UserRepository
}

func NewListListsUseCase(board port.TaskBoard, userRepo port.UserRepository) *ListListsUseCase {
	return &ListListsUseCase{board: board, userRepo: userRepo}
}

func (uc *ListListsUseCase) Execute(ctx context.Context, telegramID int64, boardID string) (*dto.ListListsOutput, error) {
	user, err := uc.userRepo.FindByTelegramID(ctx, valueobject.TelegramID(telegramID))
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}
	if !user.HasTrelloToken() {
		return nil, domainerror.ErrTrelloNotConnected
	}

	lists, err := uc.board.GetLists(ctx, user.TrelloToken(), boardID)
	if err != nil {
		return nil, fmt.Errorf("get lists: %w", err)
	}

	items := make([]dto.ListItem, len(lists))
	for i, l := range lists {
		items[i] = dto.ListItem{ID: l.ID, Name: l.Name}
	}
	return &dto.ListListsOutput{Lists: items}, nil
}
