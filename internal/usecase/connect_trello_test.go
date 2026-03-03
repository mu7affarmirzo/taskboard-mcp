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
)

func TestConnectTrello_HappyPath(t *testing.T) {
	userRepo := new(MockUserRepo)
	uc := usecase.NewConnectTrelloUseCase(userRepo)

	user := entity.NewUser(valueobject.TelegramID(100))
	userRepo.On("FindByTelegramID", mock.Anything, valueobject.TelegramID(100)).Return(user, nil)
	userRepo.On("Save", mock.Anything, mock.Anything).Return(nil)

	output, err := uc.Execute(context.Background(), dto.ConnectTrelloInput{
		TelegramID: 100,
		Token:      "trello-token-xyz",
	})

	assert.NoError(t, err)
	assert.True(t, output.Connected)
	userRepo.AssertCalled(t, "Save", mock.Anything, mock.Anything)
}

func TestConnectTrello_EmptyToken(t *testing.T) {
	userRepo := new(MockUserRepo)
	uc := usecase.NewConnectTrelloUseCase(userRepo)

	_, err := uc.Execute(context.Background(), dto.ConnectTrelloInput{
		TelegramID: 100,
		Token:      "",
	})

	assert.ErrorIs(t, err, domainerror.ErrEmptyTrelloToken)
	userRepo.AssertNotCalled(t, "FindByTelegramID")
}

func TestConnectTrello_UserNotFound(t *testing.T) {
	userRepo := new(MockUserRepo)
	uc := usecase.NewConnectTrelloUseCase(userRepo)

	userRepo.On("FindByTelegramID", mock.Anything, mock.Anything).
		Return(nil, domainerror.ErrUserNotFound)

	_, err := uc.Execute(context.Background(), dto.ConnectTrelloInput{
		TelegramID: 999,
		Token:      "some-token",
	})

	assert.Error(t, err)
	assert.ErrorIs(t, err, domainerror.ErrUserNotFound)
}

func TestConnectTrello_SaveFails(t *testing.T) {
	userRepo := new(MockUserRepo)
	uc := usecase.NewConnectTrelloUseCase(userRepo)

	user := entity.NewUser(valueobject.TelegramID(100))
	userRepo.On("FindByTelegramID", mock.Anything, mock.Anything).Return(user, nil)
	userRepo.On("Save", mock.Anything, mock.Anything).Return(errors.New("db error"))

	_, err := uc.Execute(context.Background(), dto.ConnectTrelloInput{
		TelegramID: 100,
		Token:      "some-token",
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "save user")
}
