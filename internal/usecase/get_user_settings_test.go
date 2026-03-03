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
)

func TestGetUserSettings_HappyPath(t *testing.T) {
	userRepo := new(MockUserRepo)
	uc := usecase.NewGetUserSettingsUseCase(userRepo)

	user := entity.NewUser(valueobject.TelegramID(12345))
	user.SetTrelloToken("token")
	user.SetDefaultBoard("board-1")
	user.SetDefaultList("list-1")
	userRepo.On("FindByTelegramID", mock.Anything, valueobject.TelegramID(12345)).Return(user, nil)

	output, err := uc.Execute(context.Background(), 12345)

	assert.NoError(t, err)
	assert.True(t, output.TrelloConnected)
	assert.Equal(t, "board-1", output.DefaultBoardID)
	assert.Equal(t, "list-1", output.DefaultListID)
}

func TestGetUserSettings_NoTrello(t *testing.T) {
	userRepo := new(MockUserRepo)
	uc := usecase.NewGetUserSettingsUseCase(userRepo)

	user := entity.NewUser(valueobject.TelegramID(12345))
	userRepo.On("FindByTelegramID", mock.Anything, valueobject.TelegramID(12345)).Return(user, nil)

	output, err := uc.Execute(context.Background(), 12345)

	assert.NoError(t, err)
	assert.False(t, output.TrelloConnected)
	assert.Empty(t, output.DefaultBoardID)
}

func TestGetUserSettings_UserNotFound(t *testing.T) {
	userRepo := new(MockUserRepo)
	uc := usecase.NewGetUserSettingsUseCase(userRepo)

	userRepo.On("FindByTelegramID", mock.Anything, valueobject.TelegramID(99999)).Return(nil, errors.New("not found"))

	output, err := uc.Execute(context.Background(), 99999)

	assert.Error(t, err)
	assert.Nil(t, output)
}
