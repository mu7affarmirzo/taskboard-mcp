package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"telegram-trello-bot/internal/domain/domainerror"
	"telegram-trello-bot/internal/domain/entity"
	"telegram-trello-bot/internal/domain/valueobject"
	"telegram-trello-bot/internal/usecase"
	"telegram-trello-bot/internal/usecase/port"
)

func TestListLists_HappyPath(t *testing.T) {
	board := new(MockBoard)
	userRepo := new(MockUserRepo)

	uc := usecase.NewListListsUseCase(board, userRepo)

	user := entity.NewUser(valueobject.TelegramID(100))
	user.SetTrelloToken("tok")
	userRepo.On("FindByTelegramID", mock.Anything, valueobject.TelegramID(100)).Return(user, nil)
	board.On("GetLists", mock.Anything, "tok", "board-1").Return([]port.ListInfo{
		{ID: "l1", Name: "To Do"},
		{ID: "l2", Name: "In Progress"},
		{ID: "l3", Name: "Done"},
	}, nil)

	output, err := uc.Execute(context.Background(), 100, "board-1")

	assert.NoError(t, err)
	assert.Len(t, output.Lists, 3)
	assert.Equal(t, "To Do", output.Lists[0].Name)
	assert.Equal(t, "l1", output.Lists[0].ID)
	board.AssertExpectations(t)
}

func TestListLists_UserNotFound(t *testing.T) {
	board := new(MockBoard)
	userRepo := new(MockUserRepo)

	uc := usecase.NewListListsUseCase(board, userRepo)

	userRepo.On("FindByTelegramID", mock.Anything, mock.Anything).Return(nil, domainerror.ErrUserNotFound)

	_, err := uc.Execute(context.Background(), 99999, "board-1")

	assert.Error(t, err)
	assert.ErrorIs(t, err, domainerror.ErrUserNotFound)
}

func TestListLists_GetListsFails(t *testing.T) {
	board := new(MockBoard)
	userRepo := new(MockUserRepo)

	uc := usecase.NewListListsUseCase(board, userRepo)

	user := entity.NewUser(valueobject.TelegramID(100))
	user.SetTrelloToken("tok")
	userRepo.On("FindByTelegramID", mock.Anything, mock.Anything).Return(user, nil)
	board.On("GetLists", mock.Anything, "tok", "board-1").Return(nil, errors.New("trello error"))

	_, err := uc.Execute(context.Background(), 100, "board-1")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "get lists")
}
