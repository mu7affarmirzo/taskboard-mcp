package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"telegram-trello-bot/internal/domain/entity"
	"telegram-trello-bot/internal/domain/valueobject"
	"telegram-trello-bot/internal/usecase"
	"telegram-trello-bot/internal/usecase/port"
)

func TestListBoards_HappyPath(t *testing.T) {
	board := new(MockBoard)
	userRepo := new(MockUserRepo)

	uc := usecase.NewListBoardsUseCase(board, userRepo)

	user := entity.NewUser(valueobject.TelegramID(12345))
	user.SetTrelloToken("token-abc")
	userRepo.On("FindByTelegramID", mock.Anything, valueobject.TelegramID(12345)).Return(user, nil)

	board.On("GetBoards", mock.Anything, "token-abc").Return([]port.BoardInfo{
		{ID: "b1", Name: "Work"},
		{ID: "b2", Name: "Personal"},
	}, nil)

	output, err := uc.Execute(context.Background(), 12345)

	assert.NoError(t, err)
	assert.Len(t, output.Boards, 2)
	assert.Equal(t, "b1", output.Boards[0].ID)
	assert.Equal(t, "Work", output.Boards[0].Name)
	assert.Equal(t, "b2", output.Boards[1].ID)
	assert.Equal(t, "Personal", output.Boards[1].Name)
}

func TestListBoards_UserNotFound(t *testing.T) {
	board := new(MockBoard)
	userRepo := new(MockUserRepo)

	uc := usecase.NewListBoardsUseCase(board, userRepo)

	userRepo.On("FindByTelegramID", mock.Anything, mock.Anything).Return(nil, errors.New("not found"))

	_, err := uc.Execute(context.Background(), 99999)

	assert.Error(t, err)
	board.AssertNotCalled(t, "GetBoards")
}

func TestListBoards_BoardFetchFails(t *testing.T) {
	board := new(MockBoard)
	userRepo := new(MockUserRepo)

	uc := usecase.NewListBoardsUseCase(board, userRepo)

	user := entity.NewUser(valueobject.TelegramID(12345))
	user.SetTrelloToken("token")
	userRepo.On("FindByTelegramID", mock.Anything, mock.Anything).Return(user, nil)
	board.On("GetBoards", mock.Anything, "token").Return(nil, errors.New("api error"))

	_, err := uc.Execute(context.Background(), 12345)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "get boards")
}
