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
	"telegram-trello-bot/internal/usecase/dto"
	"telegram-trello-bot/internal/usecase/port"
)

func TestCreateCardForm_HappyPath(t *testing.T) {
	board := new(MockBoard)
	userRepo := new(MockUserRepo)
	taskLog := new(MockTaskLog)
	uc := usecase.NewCreateCardFormUseCase(board, userRepo, taskLog)

	user := entity.NewUser(valueobject.TelegramID(12345))
	user.SetTrelloToken("token-abc")
	userRepo.On("FindByTelegramID", mock.Anything, valueobject.TelegramID(12345)).Return(user, nil)

	board.On("CreateCard", mock.Anything, "token-abc", mock.MatchedBy(func(p port.CreateCardParams) bool {
		return p.ListID == "list-1" && p.Title == "New Card"
	})).Return(&port.CardResult{
		CardID:  "c1",
		CardURL: "https://trello.com/c/1",
		Title:   "New Card",
	}, nil)

	taskLog.On("Log", mock.Anything, mock.Anything).Return(nil)

	output, err := uc.Execute(context.Background(), dto.CreateCardFormInput{
		TelegramID: 12345,
		ListID:     "list-1",
		Title:      "New Card",
	})

	assert.NoError(t, err)
	assert.Equal(t, "c1", output.CardID)
	assert.Equal(t, "https://trello.com/c/1", output.CardURL)
}

func TestCreateCardForm_TrelloNotConnected(t *testing.T) {
	board := new(MockBoard)
	userRepo := new(MockUserRepo)
	taskLog := new(MockTaskLog)
	uc := usecase.NewCreateCardFormUseCase(board, userRepo, taskLog)

	user := entity.NewUser(valueobject.TelegramID(12345))
	userRepo.On("FindByTelegramID", mock.Anything, valueobject.TelegramID(12345)).Return(user, nil)

	output, err := uc.Execute(context.Background(), dto.CreateCardFormInput{
		TelegramID: 12345,
		ListID:     "list-1",
		Title:      "New Card",
	})

	assert.ErrorIs(t, err, domainerror.ErrTrelloNotConnected)
	assert.Nil(t, output)
}

func TestCreateCardForm_CreateFails(t *testing.T) {
	board := new(MockBoard)
	userRepo := new(MockUserRepo)
	taskLog := new(MockTaskLog)
	uc := usecase.NewCreateCardFormUseCase(board, userRepo, taskLog)

	user := entity.NewUser(valueobject.TelegramID(12345))
	user.SetTrelloToken("token-abc")
	userRepo.On("FindByTelegramID", mock.Anything, valueobject.TelegramID(12345)).Return(user, nil)
	board.On("CreateCard", mock.Anything, "token-abc", mock.Anything).Return(nil, errors.New("api error"))

	output, err := uc.Execute(context.Background(), dto.CreateCardFormInput{
		TelegramID: 12345,
		ListID:     "list-1",
		Title:      "New Card",
	})

	assert.Error(t, err)
	assert.Nil(t, output)
}
