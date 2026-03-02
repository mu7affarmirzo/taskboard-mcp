package usecase

import (
	"context"
	"fmt"

	"telegram-trello-bot/internal/domain/domainerror"
	"telegram-trello-bot/internal/domain/valueobject"
	"telegram-trello-bot/internal/usecase/dto"
	"telegram-trello-bot/internal/usecase/port"
)

type ConfirmTaskUseCase struct {
	board    port.TaskBoard
	userRepo port.UserRepository
	taskLog  port.TaskLogRepository
}

func NewConfirmTaskUseCase(
	board port.TaskBoard,
	userRepo port.UserRepository,
	taskLog port.TaskLogRepository,
) *ConfirmTaskUseCase {
	return &ConfirmTaskUseCase{
		board:    board,
		userRepo: userRepo,
		taskLog:  taskLog,
	}
}

func (uc *ConfirmTaskUseCase) Execute(
	ctx context.Context,
	input dto.ConfirmTaskInput,
) (*dto.CreateTaskOutput, error) {
	user, err := uc.userRepo.FindByTelegramID(ctx, valueobject.TelegramID(input.TelegramID))
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}
	if !user.HasBoardConfigured() {
		return nil, domainerror.ErrBoardNotSet
	}
	if !user.HasListConfigured() {
		return nil, domainerror.ErrListNotSet
	}

	// Resolve label names to Trello label IDs
	var labelIDs []string
	if len(input.Labels) > 0 {
		ids, err := uc.board.MatchLabels(ctx, user.TrelloToken(), user.DefaultBoard(), input.Labels)
		if err != nil {
			return nil, fmt.Errorf("match labels: %w", err)
		}
		labelIDs = ids
	}

	// Map to card params
	var dueStr *string
	if input.DueDate != nil {
		s := input.DueDate.Format("2006-01-02T15:04:05Z")
		dueStr = &s
	}

	priority, _ := valueobject.NewPriority(input.Priority)
	position := "bottom"
	if priority == valueobject.PriorityHigh {
		position = "top"
	}

	params := port.CreateCardParams{
		ListID:      user.DefaultList(),
		Title:       input.Title,
		Description: input.Description,
		DueDate:     dueStr,
		LabelIDs:    labelIDs,
		Position:    position,
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

	return &dto.CreateTaskOutput{
		CardURL:   result.CardURL,
		TaskTitle: input.Title,
		DueDate:   input.DueDate,
		Priority:  input.Priority,
		Labels:    input.Labels,
	}, nil
}
