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

func TestConfirmTask_HappyPath(t *testing.T) {
	board := new(MockBoard)
	userRepo := new(MockUserRepo)
	taskLog := new(MockTaskLog)

	memberResolver := new(MockMemberResolver)
	uc := usecase.NewConfirmTaskUseCase(board, memberResolver, userRepo, taskLog)

	user := newConfiguredUser()
	userRepo.On("FindByTelegramID", mock.Anything, valueobject.TelegramID(12345)).Return(user, nil)

	board.On("MatchLabels", mock.Anything, "trello-token-xyz", "board-1", []string{"backend"}).
		Return([]string{"label-id-1"}, nil)

	board.On("CreateCard", mock.Anything, "trello-token-xyz", mock.MatchedBy(func(p port.CreateCardParams) bool {
		return p.Title == "Fix bug" && p.Position == "top" && p.ListID == "list-1" &&
			len(p.LabelIDs) == 1 && p.LabelIDs[0] == "label-id-1"
	})).Return(&port.CardResult{
		CardID: "card-123", CardURL: "https://trello.com/c/abc123",
	}, nil)

	taskLog.On("Log", mock.Anything, mock.Anything).Return(nil)

	due := time.Date(2025, 3, 7, 0, 0, 0, 0, time.UTC)
	output, err := uc.Execute(context.Background(), dto.ConfirmTaskInput{
		TelegramID:  12345,
		Title:       "Fix bug",
		Description: "Payment module",
		DueDate:     &due,
		Priority:    "high",
		Labels:      []string{"backend"},
	})

	assert.NoError(t, err)
	assert.Equal(t, "Fix bug", output.TaskTitle)
	assert.Equal(t, "https://trello.com/c/abc123", output.CardURL)
	assert.Equal(t, "high", output.Priority)
	board.AssertExpectations(t)
	taskLog.AssertExpectations(t)
}

func TestConfirmTask_TrelloNotConnected(t *testing.T) {
	board := new(MockBoard)
	userRepo := new(MockUserRepo)
	taskLog := new(MockTaskLog)

	memberResolver := new(MockMemberResolver)
	uc := usecase.NewConfirmTaskUseCase(board, memberResolver, userRepo, taskLog)

	user := entity.NewUser(valueobject.TelegramID(12345))
	userRepo.On("FindByTelegramID", mock.Anything, mock.Anything).Return(user, nil)

	_, err := uc.Execute(context.Background(), dto.ConfirmTaskInput{
		TelegramID: 12345, Title: "task",
	})

	assert.ErrorIs(t, err, domainerror.ErrTrelloNotConnected)
}

func TestConfirmTask_BoardNotSet(t *testing.T) {
	board := new(MockBoard)
	userRepo := new(MockUserRepo)
	taskLog := new(MockTaskLog)

	memberResolver := new(MockMemberResolver)
	uc := usecase.NewConfirmTaskUseCase(board, memberResolver, userRepo, taskLog)

	user := entity.NewUser(valueobject.TelegramID(12345))
	user.SetTrelloToken("tok")
	userRepo.On("FindByTelegramID", mock.Anything, mock.Anything).Return(user, nil)

	_, err := uc.Execute(context.Background(), dto.ConfirmTaskInput{
		TelegramID: 12345, Title: "task",
	})

	assert.ErrorIs(t, err, domainerror.ErrBoardNotSet)
}

func TestConfirmTask_CreateCardFails(t *testing.T) {
	board := new(MockBoard)
	userRepo := new(MockUserRepo)
	taskLog := new(MockTaskLog)

	memberResolver := new(MockMemberResolver)
	uc := usecase.NewConfirmTaskUseCase(board, memberResolver, userRepo, taskLog)

	user := newConfiguredUser()
	userRepo.On("FindByTelegramID", mock.Anything, mock.Anything).Return(user, nil)
	board.On("CreateCard", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("trello down"))

	_, err := uc.Execute(context.Background(), dto.ConfirmTaskInput{
		TelegramID: 12345, Title: "task",
	})

	assert.ErrorIs(t, err, domainerror.ErrCardCreation)
}

func TestConfirmTask_NormalPriority_BottomPosition(t *testing.T) {
	board := new(MockBoard)
	userRepo := new(MockUserRepo)
	taskLog := new(MockTaskLog)

	memberResolver := new(MockMemberResolver)
	uc := usecase.NewConfirmTaskUseCase(board, memberResolver, userRepo, taskLog)

	user := newConfiguredUser()
	userRepo.On("FindByTelegramID", mock.Anything, mock.Anything).Return(user, nil)

	board.On("CreateCard", mock.Anything, mock.Anything, mock.MatchedBy(func(p port.CreateCardParams) bool {
		return p.Position == "bottom"
	})).Return(&port.CardResult{CardID: "c1", CardURL: "url"}, nil)

	taskLog.On("Log", mock.Anything, mock.Anything).Return(nil)

	output, err := uc.Execute(context.Background(), dto.ConfirmTaskInput{
		TelegramID: 12345, Title: "Normal task", Priority: "medium",
	})

	assert.NoError(t, err)
	assert.Equal(t, "medium", output.Priority)
	board.AssertExpectations(t)
}
