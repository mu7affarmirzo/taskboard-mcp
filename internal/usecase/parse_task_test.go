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
)

func TestParseTask_HappyPath(t *testing.T) {
	parser := new(MockParser)
	userRepo := new(MockUserRepo)

	uc := usecase.NewParseTaskUseCase(parser, userRepo)

	user := newConfiguredUser()
	userRepo.On("FindByTelegramID", mock.Anything, valueobject.TelegramID(12345)).Return(user, nil)

	due := time.Date(2025, 3, 7, 0, 0, 0, 0, time.UTC)
	task, _ := entity.NewTask("Fix bug",
		entity.WithDescription("Payment module"),
		entity.WithPriority(valueobject.PriorityHigh),
		entity.WithDueDate(due),
		entity.WithLabels([]string{"backend"}),
		entity.WithChecklist([]string{"step 1"}),
	)
	parser.On("Parse", mock.Anything, "Fix bug urgent").Return(task, nil)

	output, err := uc.Execute(context.Background(), dto.CreateTaskInput{
		TelegramID: 12345,
		RawMessage: "Fix bug urgent",
	})

	assert.NoError(t, err)
	assert.Equal(t, "Fix bug", output.TaskTitle)
	assert.Equal(t, "Payment module", output.Description)
	assert.Equal(t, "high", output.Priority)
	assert.Equal(t, &due, output.DueDate)
	assert.Equal(t, []string{"backend"}, output.Labels)
	assert.Equal(t, []string{"step 1"}, output.Checklist)
	assert.Equal(t, "board-1", output.BoardID)
	assert.Equal(t, "list-1", output.ListID)
}

func TestParseTask_TrelloNotConnected(t *testing.T) {
	parser := new(MockParser)
	userRepo := new(MockUserRepo)

	uc := usecase.NewParseTaskUseCase(parser, userRepo)

	user := entity.NewUser(valueobject.TelegramID(12345))
	userRepo.On("FindByTelegramID", mock.Anything, mock.Anything).Return(user, nil)

	_, err := uc.Execute(context.Background(), dto.CreateTaskInput{
		TelegramID: 12345, RawMessage: "task",
	})

	assert.ErrorIs(t, err, domainerror.ErrTrelloNotConnected)
	parser.AssertNotCalled(t, "Parse")
}

func TestParseTask_BoardNotSet(t *testing.T) {
	parser := new(MockParser)
	userRepo := new(MockUserRepo)

	uc := usecase.NewParseTaskUseCase(parser, userRepo)

	user := entity.NewUser(valueobject.TelegramID(12345))
	user.SetTrelloToken("tok")
	userRepo.On("FindByTelegramID", mock.Anything, mock.Anything).Return(user, nil)

	_, err := uc.Execute(context.Background(), dto.CreateTaskInput{
		TelegramID: 12345, RawMessage: "task",
	})

	assert.ErrorIs(t, err, domainerror.ErrBoardNotSet)
	parser.AssertNotCalled(t, "Parse")
}

func TestParseTask_ParseFails(t *testing.T) {
	parser := new(MockParser)
	userRepo := new(MockUserRepo)

	uc := usecase.NewParseTaskUseCase(parser, userRepo)

	user := newConfiguredUser()
	userRepo.On("FindByTelegramID", mock.Anything, mock.Anything).Return(user, nil)
	parser.On("Parse", mock.Anything, mock.Anything).Return(nil, errors.New("parse error"))

	_, err := uc.Execute(context.Background(), dto.CreateTaskInput{
		TelegramID: 12345, RawMessage: "???",
	})

	assert.ErrorIs(t, err, domainerror.ErrParsingFailed)
}

func TestParseTask_UserNotFound(t *testing.T) {
	parser := new(MockParser)
	userRepo := new(MockUserRepo)

	uc := usecase.NewParseTaskUseCase(parser, userRepo)

	userRepo.On("FindByTelegramID", mock.Anything, mock.Anything).Return(nil, domainerror.ErrUserNotFound)

	_, err := uc.Execute(context.Background(), dto.CreateTaskInput{
		TelegramID: 99999, RawMessage: "task",
	})

	assert.ErrorIs(t, err, domainerror.ErrUserNotFound)
}
