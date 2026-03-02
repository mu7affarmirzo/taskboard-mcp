package usecase

import (
	"context"
	"fmt"

	"telegram-trello-bot/internal/domain/valueobject"
	"telegram-trello-bot/internal/usecase/dto"
	"telegram-trello-bot/internal/usecase/port"
)

type ListBoardsUseCase struct {
	board    port.TaskBoard
	userRepo port.UserRepository
}

func NewListBoardsUseCase(board port.TaskBoard, userRepo port.UserRepository) *ListBoardsUseCase {
	return &ListBoardsUseCase{board: board, userRepo: userRepo}
}

func (uc *ListBoardsUseCase) Execute(ctx context.Context, telegramID int64) (*dto.ListBoardsOutput, error) {
	user, err := uc.userRepo.FindByTelegramID(ctx, valueobject.TelegramID(telegramID))
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}

	boards, err := uc.board.GetBoards(ctx, user.TrelloToken())
	if err != nil {
		return nil, fmt.Errorf("get boards: %w", err)
	}

	items := make([]dto.BoardItem, len(boards))
	for i, b := range boards {
		items[i] = dto.BoardItem{ID: b.ID, Name: b.Name}
	}
	return &dto.ListBoardsOutput{Boards: items}, nil
}
