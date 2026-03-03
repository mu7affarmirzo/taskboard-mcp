package usecase

import (
	"context"
	"fmt"

	"telegram-trello-bot/internal/domain/domainerror"
	"telegram-trello-bot/internal/domain/valueobject"
	"telegram-trello-bot/internal/usecase/dto"
	"telegram-trello-bot/internal/usecase/port"
)

type CreateCardFormUseCase struct {
	board    port.TaskBoard
	userRepo port.UserRepository
	taskLog  port.TaskLogRepository
}

func NewCreateCardFormUseCase(
	board port.TaskBoard,
	userRepo port.UserRepository,
	taskLog port.TaskLogRepository,
) *CreateCardFormUseCase {
	return &CreateCardFormUseCase{
		board:    board,
		userRepo: userRepo,
		taskLog:  taskLog,
	}
}

func (uc *CreateCardFormUseCase) Execute(ctx context.Context, input dto.CreateCardFormInput) (*dto.CreateCardFormOutput, error) {
	user, err := uc.userRepo.FindByTelegramID(ctx, valueobject.TelegramID(input.TelegramID))
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}
	if !user.HasTrelloToken() {
		return nil, domainerror.ErrTrelloNotConnected
	}

	var dueDate *string
	if input.DueDate != "" {
		dueDate = &input.DueDate
	}

	params := port.CreateCardParams{
		ListID:      input.ListID,
		Title:       input.Title,
		Description: input.Description,
		DueDate:     dueDate,
		LabelIDs:    input.LabelIDs,
		MemberIDs:   input.MemberIDs,
		Position:    "bottom",
	}

	result, err := uc.board.CreateCard(ctx, user.TrelloToken(), params)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domainerror.ErrCardCreation, err)
	}

	_ = uc.taskLog.Log(ctx, port.TaskLogEntry{
		TelegramID: input.TelegramID,
		Message:    input.Title,
		CardID:     result.CardID,
	})

	return &dto.CreateCardFormOutput{
		CardID:  result.CardID,
		CardURL: result.CardURL,
		Title:   result.Title,
	}, nil
}
