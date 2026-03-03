package usecase

import (
	"context"
	"fmt"

	"telegram-trello-bot/internal/domain/domainerror"
	"telegram-trello-bot/internal/domain/valueobject"
	"telegram-trello-bot/internal/usecase/dto"
	"telegram-trello-bot/internal/usecase/port"
)

type ParseTaskUseCase struct {
	parser   port.TaskParser
	userRepo port.UserRepository
}

func NewParseTaskUseCase(
	parser port.TaskParser,
	userRepo port.UserRepository,
) *ParseTaskUseCase {
	return &ParseTaskUseCase{
		parser:   parser,
		userRepo: userRepo,
	}
}

func (uc *ParseTaskUseCase) Execute(
	ctx context.Context,
	input dto.CreateTaskInput,
) (*dto.ParseTaskOutput, error) {
	user, err := uc.userRepo.FindByTelegramID(ctx, valueobject.TelegramID(input.TelegramID))
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}
	if !user.HasTrelloToken() {
		return nil, domainerror.ErrTrelloNotConnected
	}
	if !user.HasBoardConfigured() {
		return nil, domainerror.ErrBoardNotSet
	}
	if !user.HasListConfigured() {
		return nil, domainerror.ErrListNotSet
	}

	task, err := uc.parser.Parse(ctx, input.RawMessage)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domainerror.ErrParsingFailed, err)
	}

	return &dto.ParseTaskOutput{
		TaskTitle:   task.Title(),
		Description: task.Description(),
		DueDate:     task.DueDate(),
		Priority:    string(task.Priority()),
		Labels:      task.Labels(),
		Checklist:   task.Checklist(),
		Members:     task.Members(),
		BoardID:     user.DefaultBoard(),
		ListID:      user.DefaultList(),
	}, nil
}
