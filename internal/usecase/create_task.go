package usecase

import (
	"context"
	"fmt"

	"telegram-trello-bot/internal/domain/domainerror"
	"telegram-trello-bot/internal/domain/valueobject"
	"telegram-trello-bot/internal/usecase/dto"
	"telegram-trello-bot/internal/usecase/port"
)

type CreateTaskUseCase struct {
	parser         port.TaskParser
	board          port.TaskBoard
	memberResolver port.MemberResolver
	userRepo       port.UserRepository
	taskLog        port.TaskLogRepository
}

func NewCreateTaskUseCase(
	parser port.TaskParser,
	board port.TaskBoard,
	memberResolver port.MemberResolver,
	userRepo port.UserRepository,
	taskLog port.TaskLogRepository,
) *CreateTaskUseCase {
	return &CreateTaskUseCase{
		parser:         parser,
		board:          board,
		memberResolver: memberResolver,
		userRepo:       userRepo,
		taskLog:        taskLog,
	}
}

func (uc *CreateTaskUseCase) Execute(
	ctx context.Context,
	input dto.CreateTaskInput,
) (*dto.CreateTaskOutput, error) {
	// 1. Load user preferences
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

	// 2. Parse message into domain Task entity
	task, err := uc.parser.Parse(ctx, input.RawMessage)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domainerror.ErrParsingFailed, err)
	}

	// 3. Resolve label names to Trello label IDs
	var labelIDs []string
	if len(task.Labels()) > 0 {
		ids, err := uc.board.MatchLabels(ctx, user.TrelloToken(), user.DefaultBoard(), task.Labels())
		if err != nil {
			return nil, fmt.Errorf("match labels: %w", err)
		}
		labelIDs = ids
	}

	// 3b. Resolve member names to Trello member IDs
	var memberIDs []string
	if len(task.Members()) > 0 {
		ids, err := uc.memberResolver.MatchMembers(ctx, user.TrelloToken(), user.DefaultBoard(), task.Members())
		if err != nil {
			return nil, fmt.Errorf("match members: %w", err)
		}
		memberIDs = ids
	}

	// 4. Map domain entity to board card params
	var dueStr *string
	if task.DueDate() != nil {
		s := task.DueDate().Format("2006-01-02T15:04:05Z")
		dueStr = &s
	}

	position := "bottom"
	if task.IsHighPriority() {
		position = "top"
	}

	params := port.CreateCardParams{
		ListID:      user.DefaultList(),
		Title:       task.Title(),
		Description: task.Description(),
		DueDate:     dueStr,
		LabelIDs:    labelIDs,
		MemberIDs:   memberIDs,
		Position:    position,
	}

	// 5. Create card on board
	result, err := uc.board.CreateCard(ctx, user.TrelloToken(), params)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domainerror.ErrCardCreation, err)
	}

	// 6. Log task creation
	_ = uc.taskLog.Log(ctx, port.TaskLogEntry{
		TelegramID: input.TelegramID,
		Message:    input.RawMessage,
		CardID:     result.CardID,
	})

	// 7. Return output DTO
	return &dto.CreateTaskOutput{
		CardURL:   result.CardURL,
		TaskTitle: task.Title(),
		DueDate:   task.DueDate(),
		Priority:  string(task.Priority()),
		Labels:    task.Labels(),
		Members:   task.Members(),
	}, nil
}
