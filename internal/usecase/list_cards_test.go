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

func TestListCards_HappyPath(t *testing.T) {
	board := new(MockBoard)
	userRepo := new(MockUserRepo)
	uc := usecase.NewListCardsUseCase(board, userRepo)

	user := entity.NewUser(valueobject.TelegramID(12345))
	user.SetTrelloToken("token-abc")
	userRepo.On("FindByTelegramID", mock.Anything, valueobject.TelegramID(12345)).Return(user, nil)

	board.On("GetCards", mock.Anything, "token-abc", "list-1").Return([]port.CardResult{
		{CardID: "c1", Title: "Task 1", CardURL: "https://trello.com/c/1", ListID: "list-1"},
		{CardID: "c2", Title: "Task 2", CardURL: "https://trello.com/c/2", ListID: "list-1"},
	}, nil)

	output, err := uc.Execute(context.Background(), 12345, "list-1")

	assert.NoError(t, err)
	assert.Len(t, output.Cards, 2)
	assert.Equal(t, "Task 1", output.Cards[0].Title)
}

func TestListCards_TrelloNotConnected(t *testing.T) {
	board := new(MockBoard)
	userRepo := new(MockUserRepo)
	uc := usecase.NewListCardsUseCase(board, userRepo)

	user := entity.NewUser(valueobject.TelegramID(12345))
	userRepo.On("FindByTelegramID", mock.Anything, valueobject.TelegramID(12345)).Return(user, nil)

	output, err := uc.Execute(context.Background(), 12345, "list-1")

	assert.ErrorIs(t, err, domainerror.ErrTrelloNotConnected)
	assert.Nil(t, output)
}

func TestListCards_FetchFails(t *testing.T) {
	board := new(MockBoard)
	userRepo := new(MockUserRepo)
	uc := usecase.NewListCardsUseCase(board, userRepo)

	user := entity.NewUser(valueobject.TelegramID(12345))
	user.SetTrelloToken("token-abc")
	userRepo.On("FindByTelegramID", mock.Anything, valueobject.TelegramID(12345)).Return(user, nil)
	board.On("GetCards", mock.Anything, "token-abc", "list-1").Return(nil, errors.New("api error"))

	output, err := uc.Execute(context.Background(), 12345, "list-1")

	assert.Error(t, err)
	assert.Nil(t, output)
}
