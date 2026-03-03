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

func TestRegisterUser_NewUser(t *testing.T) {
	userRepo := new(MockUserRepo)
	uc := usecase.NewRegisterUserUseCase(userRepo, "test-api-key")

	userRepo.On("FindByTelegramID", mock.Anything, valueobject.TelegramID(100)).
		Return(nil, domainerror.ErrUserNotFound)
	userRepo.On("Save", mock.Anything, mock.Anything).Return(nil)

	output, err := uc.Execute(context.Background(), dto.RegisterUserInput{TelegramID: 100})

	assert.NoError(t, err)
	assert.True(t, output.IsNewUser)
	assert.Contains(t, output.TrelloAuthURL, "test-api-key")
	assert.Contains(t, output.TrelloAuthURL, "trello.com/1/authorize")
	userRepo.AssertCalled(t, "Save", mock.Anything, mock.Anything)
}

func TestRegisterUser_ExistingUser(t *testing.T) {
	userRepo := new(MockUserRepo)
	uc := usecase.NewRegisterUserUseCase(userRepo, "test-api-key")

	user := entity.NewUser(valueobject.TelegramID(100))
	userRepo.On("FindByTelegramID", mock.Anything, valueobject.TelegramID(100)).
		Return(user, nil)

	output, err := uc.Execute(context.Background(), dto.RegisterUserInput{TelegramID: 100})

	assert.NoError(t, err)
	assert.False(t, output.IsNewUser)
	assert.Contains(t, output.TrelloAuthURL, "test-api-key")
	userRepo.AssertNotCalled(t, "Save")
}

func TestRegisterUser_SaveFails(t *testing.T) {
	userRepo := new(MockUserRepo)
	uc := usecase.NewRegisterUserUseCase(userRepo, "test-api-key")

	userRepo.On("FindByTelegramID", mock.Anything, mock.Anything).
		Return(nil, domainerror.ErrUserNotFound)
	userRepo.On("Save", mock.Anything, mock.Anything).Return(errors.New("db error"))

	_, err := uc.Execute(context.Background(), dto.RegisterUserInput{TelegramID: 100})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "save user")
}

func TestRegisterUser_FindError(t *testing.T) {
	userRepo := new(MockUserRepo)
	uc := usecase.NewRegisterUserUseCase(userRepo, "test-api-key")

	userRepo.On("FindByTelegramID", mock.Anything, mock.Anything).
		Return(nil, errors.New("db connection failed"))

	_, err := uc.Execute(context.Background(), dto.RegisterUserInput{TelegramID: 100})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "find user")
}
