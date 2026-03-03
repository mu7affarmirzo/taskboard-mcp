package usecase

import (
	"context"
	"fmt"

	"telegram-trello-bot/internal/domain/domainerror"
	"telegram-trello-bot/internal/domain/valueobject"
	"telegram-trello-bot/internal/usecase/dto"
	"telegram-trello-bot/internal/usecase/port"
)

type ListLabelsUseCase struct {
	board    port.TaskBoard
	userRepo port.UserRepository
}

func NewListLabelsUseCase(board port.TaskBoard, userRepo port.UserRepository) *ListLabelsUseCase {
	return &ListLabelsUseCase{board: board, userRepo: userRepo}
}

func (uc *ListLabelsUseCase) Execute(ctx context.Context, telegramID int64, boardID string) (*dto.ListLabelsOutput, error) {
	user, err := uc.userRepo.FindByTelegramID(ctx, valueobject.TelegramID(telegramID))
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}
	if !user.HasTrelloToken() {
		return nil, domainerror.ErrTrelloNotConnected
	}

	labels, err := uc.board.GetLabels(ctx, user.TrelloToken(), boardID)
	if err != nil {
		return nil, fmt.Errorf("get labels: %w", err)
	}

	items := make([]dto.LabelItem, len(labels))
	for i, l := range labels {
		items[i] = dto.LabelItem{ID: l.ID, Name: l.Name, Color: l.Color}
	}
	return &dto.ListLabelsOutput{Labels: items}, nil
}
