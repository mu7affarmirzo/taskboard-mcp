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
)

func TestDeleteCard_HappyPath(t *testing.T) {
	cardMgr := new(MockCardManager)
	userRepo := new(MockUserRepo)
	uc := usecase.NewDeleteCardUseCase(cardMgr, userRepo)

	user := entity.NewUser(valueobject.TelegramID(12345))
	user.SetTrelloToken("token-abc")
	userRepo.On("FindByTelegramID", mock.Anything, valueobject.TelegramID(12345)).Return(user, nil)
	cardMgr.On("DeleteCard", mock.Anything, "token-abc", "card-1").Return(nil)

	err := uc.Execute(context.Background(), 12345, "card-1")

	assert.NoError(t, err)
}

func TestDeleteCard_TrelloNotConnected(t *testing.T) {
	cardMgr := new(MockCardManager)
	userRepo := new(MockUserRepo)
	uc := usecase.NewDeleteCardUseCase(cardMgr, userRepo)

	user := entity.NewUser(valueobject.TelegramID(12345))
	userRepo.On("FindByTelegramID", mock.Anything, valueobject.TelegramID(12345)).Return(user, nil)

	err := uc.Execute(context.Background(), 12345, "card-1")

	assert.ErrorIs(t, err, domainerror.ErrTrelloNotConnected)
}

func TestDeleteCard_Fails(t *testing.T) {
	cardMgr := new(MockCardManager)
	userRepo := new(MockUserRepo)
	uc := usecase.NewDeleteCardUseCase(cardMgr, userRepo)

	user := entity.NewUser(valueobject.TelegramID(12345))
	user.SetTrelloToken("token-abc")
	userRepo.On("FindByTelegramID", mock.Anything, valueobject.TelegramID(12345)).Return(user, nil)
	cardMgr.On("DeleteCard", mock.Anything, "token-abc", "card-1").Return(errors.New("api error"))

	err := uc.Execute(context.Background(), 12345, "card-1")

	assert.Error(t, err)
}
