package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"telegram-trello-bot/internal/domain/domainerror"
	"telegram-trello-bot/internal/domain/entity"
	"telegram-trello-bot/internal/domain/valueobject"
	"telegram-trello-bot/internal/usecase"
	"telegram-trello-bot/internal/usecase/dto"
	"telegram-trello-bot/internal/usecase/port"
)

func newConfiguredUser() *entity.User {
	user := entity.NewUser(valueobject.TelegramID(12345))
	user.SetDefaultBoard("board-1")
	user.SetDefaultList("list-1")
	user.SetTrelloToken("trello-token-xyz")
	return user
}

func TestCreateTask_HappyPath(t *testing.T) {
	parser := new(MockParser)
	board := new(MockBoard)
	userRepo := new(MockUserRepo)
	taskLog := new(MockTaskLog)

	memberResolver := new(MockMemberResolver)
	uc := usecase.NewCreateTaskUseCase(parser, board, memberResolver, userRepo, taskLog)

	user := newConfiguredUser()
	userRepo.On("FindByTelegramID", mock.Anything, valueobject.TelegramID(12345)).Return(user, nil)

	due := time.Date(2025, 3, 7, 0, 0, 0, 0, time.UTC)
	task, _ := entity.NewTask("Fix payment bug",
		entity.WithPriority(valueobject.PriorityHigh),
		entity.WithDueDate(due),
		entity.WithLabels([]string{"backend"}),
	)
	parser.On("Parse", mock.Anything, mock.Anything).Return(task, nil)

	board.On("MatchLabels", mock.Anything, "trello-token-xyz", "board-1", []string{"backend"}).
		Return([]string{"label-id-1"}, nil)

	board.On("CreateCard", mock.Anything, "trello-token-xyz", mock.MatchedBy(func(p port.CreateCardParams) bool {
		return p.Title == "Fix payment bug" && p.Position == "top" && p.ListID == "list-1" &&
			len(p.LabelIDs) == 1 && p.LabelIDs[0] == "label-id-1"
	})).Return(&port.CardResult{
		CardID: "card-123", CardURL: "https://trello.com/c/abc123",
	}, nil)

	taskLog.On("Log", mock.Anything, mock.Anything).Return(nil)

	output, err := uc.Execute(context.Background(), dto.CreateTaskInput{
		TelegramID: 12345,
		RawMessage: "Fix payment bug, urgent, due Friday #backend",
	})

	assert.NoError(t, err)
	assert.Equal(t, "Fix payment bug", output.TaskTitle)
	assert.Equal(t, "https://trello.com/c/abc123", output.CardURL)
	assert.Equal(t, "high", output.Priority)
	assert.Equal(t, []string{"backend"}, output.Labels)
	board.AssertExpectations(t)
	taskLog.AssertExpectations(t)
}

func TestCreateTask_TrelloNotConnected(t *testing.T) {
	parser := new(MockParser)
	board := new(MockBoard)
	userRepo := new(MockUserRepo)
	taskLog := new(MockTaskLog)

	memberResolver := new(MockMemberResolver)
	uc := usecase.NewCreateTaskUseCase(parser, board, memberResolver, userRepo, taskLog)

	user := entity.NewUser(valueobject.TelegramID(12345))
	userRepo.On("FindByTelegramID", mock.Anything, mock.Anything).Return(user, nil)

	_, err := uc.Execute(context.Background(), dto.CreateTaskInput{
		TelegramID: 12345, RawMessage: "Some task",
	})

	assert.ErrorIs(t, err, domainerror.ErrTrelloNotConnected)
	parser.AssertNotCalled(t, "Parse")
}

func TestCreateTask_BoardNotSet(t *testing.T) {
	parser := new(MockParser)
	board := new(MockBoard)
	userRepo := new(MockUserRepo)
	taskLog := new(MockTaskLog)

	memberResolver := new(MockMemberResolver)
	uc := usecase.NewCreateTaskUseCase(parser, board, memberResolver, userRepo, taskLog)

	user := entity.NewUser(valueobject.TelegramID(12345))
	user.SetTrelloToken("tok")
	userRepo.On("FindByTelegramID", mock.Anything, mock.Anything).Return(user, nil)

	_, err := uc.Execute(context.Background(), dto.CreateTaskInput{
		TelegramID: 12345, RawMessage: "Some task",
	})

	assert.ErrorIs(t, err, domainerror.ErrBoardNotSet)
	parser.AssertNotCalled(t, "Parse")
}

func TestCreateTask_ListNotSet(t *testing.T) {
	parser := new(MockParser)
	board := new(MockBoard)
	userRepo := new(MockUserRepo)
	taskLog := new(MockTaskLog)

	memberResolver := new(MockMemberResolver)
	uc := usecase.NewCreateTaskUseCase(parser, board, memberResolver, userRepo, taskLog)

	user := entity.NewUser(valueobject.TelegramID(12345))
	user.SetTrelloToken("tok")
	user.SetDefaultBoard("board-1")
	userRepo.On("FindByTelegramID", mock.Anything, mock.Anything).Return(user, nil)

	_, err := uc.Execute(context.Background(), dto.CreateTaskInput{
		TelegramID: 12345, RawMessage: "Some task",
	})

	assert.ErrorIs(t, err, domainerror.ErrListNotSet)
	parser.AssertNotCalled(t, "Parse")
}

func TestCreateTask_ParseFails(t *testing.T) {
	parser := new(MockParser)
	board := new(MockBoard)
	userRepo := new(MockUserRepo)
	taskLog := new(MockTaskLog)

	memberResolver := new(MockMemberResolver)
	uc := usecase.NewCreateTaskUseCase(parser, board, memberResolver, userRepo, taskLog)

	user := newConfiguredUser()
	userRepo.On("FindByTelegramID", mock.Anything, mock.Anything).Return(user, nil)
	parser.On("Parse", mock.Anything, mock.Anything).Return(nil, errors.New("parse error"))

	_, err := uc.Execute(context.Background(), dto.CreateTaskInput{
		TelegramID: 12345, RawMessage: "???",
	})

	assert.ErrorIs(t, err, domainerror.ErrParsingFailed)
}

func TestCreateTask_CreateCardFails(t *testing.T) {
	parser := new(MockParser)
	board := new(MockBoard)
	userRepo := new(MockUserRepo)
	taskLog := new(MockTaskLog)

	memberResolver := new(MockMemberResolver)
	uc := usecase.NewCreateTaskUseCase(parser, board, memberResolver, userRepo, taskLog)

	user := newConfiguredUser()
	userRepo.On("FindByTelegramID", mock.Anything, mock.Anything).Return(user, nil)

	task, _ := entity.NewTask("Some task")
	parser.On("Parse", mock.Anything, mock.Anything).Return(task, nil)
	board.On("CreateCard", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("trello down"))

	_, err := uc.Execute(context.Background(), dto.CreateTaskInput{
		TelegramID: 12345, RawMessage: "Some task",
	})

	assert.ErrorIs(t, err, domainerror.ErrCardCreation)
}

func TestCreateTask_UserNotFound(t *testing.T) {
	parser := new(MockParser)
	board := new(MockBoard)
	userRepo := new(MockUserRepo)
	taskLog := new(MockTaskLog)

	memberResolver := new(MockMemberResolver)
	uc := usecase.NewCreateTaskUseCase(parser, board, memberResolver, userRepo, taskLog)

	userRepo.On("FindByTelegramID", mock.Anything, mock.Anything).Return(nil, domainerror.ErrUserNotFound)

	_, err := uc.Execute(context.Background(), dto.CreateTaskInput{
		TelegramID: 99999, RawMessage: "task",
	})

	assert.Error(t, err)
	assert.ErrorIs(t, err, domainerror.ErrUserNotFound)
}

func TestCreateTask_NormalPriority_BottomPosition(t *testing.T) {
	parser := new(MockParser)
	board := new(MockBoard)
	userRepo := new(MockUserRepo)
	taskLog := new(MockTaskLog)

	memberResolver := new(MockMemberResolver)
	uc := usecase.NewCreateTaskUseCase(parser, board, memberResolver, userRepo, taskLog)

	user := newConfiguredUser()
	userRepo.On("FindByTelegramID", mock.Anything, mock.Anything).Return(user, nil)

	task, _ := entity.NewTask("Normal task", entity.WithPriority(valueobject.PriorityMedium))
	parser.On("Parse", mock.Anything, mock.Anything).Return(task, nil)

	board.On("CreateCard", mock.Anything, mock.Anything, mock.MatchedBy(func(p port.CreateCardParams) bool {
		return p.Position == "bottom"
	})).Return(&port.CardResult{CardID: "c1", CardURL: "url"}, nil)

	taskLog.On("Log", mock.Anything, mock.Anything).Return(nil)

	output, err := uc.Execute(context.Background(), dto.CreateTaskInput{
		TelegramID: 12345, RawMessage: "Normal task",
	})

	assert.NoError(t, err)
	assert.Equal(t, "medium", output.Priority)
	board.AssertExpectations(t)
}
