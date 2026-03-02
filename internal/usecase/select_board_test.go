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

func TestSelectBoard_HappyPath(t *testing.T) {
	userRepo := new(MockUserRepo)

	uc := usecase.NewSelectBoardUseCase(userRepo)

	user := entity.NewUser(valueobject.TelegramID(12345))
	userRepo.On("FindByTelegramID", mock.Anything, valueobject.TelegramID(12345)).Return(user, nil)
	userRepo.On("Save", mock.Anything, mock.MatchedBy(func(u *entity.User) bool {
		return u.DefaultBoard() == "board-abc"
	})).Return(nil)

	err := uc.Execute(context.Background(), 12345, "board-abc")

	assert.NoError(t, err)
	userRepo.AssertExpectations(t)
}

func TestSelectBoard_UserNotFound(t *testing.T) {
	userRepo := new(MockUserRepo)

	uc := usecase.NewSelectBoardUseCase(userRepo)

	userRepo.On("FindByTelegramID", mock.Anything, mock.Anything).Return(nil, errors.New("not found"))

	err := uc.Execute(context.Background(), 99999, "board-abc")

	assert.Error(t, err)
	userRepo.AssertNotCalled(t, "Save")
}

func TestSelectBoard_SaveFails(t *testing.T) {
	userRepo := new(MockUserRepo)

	uc := usecase.NewSelectBoardUseCase(userRepo)

	user := entity.NewUser(valueobject.TelegramID(12345))
	userRepo.On("FindByTelegramID", mock.Anything, mock.Anything).Return(user, nil)
	userRepo.On("Save", mock.Anything, mock.Anything).Return(errors.New("db error"))

	err := uc.Execute(context.Background(), 12345, "board-abc")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}
