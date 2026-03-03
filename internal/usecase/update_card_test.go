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

func TestUpdateCard_HappyPath(t *testing.T) {
	cardMgr := new(MockCardManager)
	userRepo := new(MockUserRepo)
	uc := usecase.NewUpdateCardUseCase(cardMgr, userRepo)

	user := entity.NewUser(valueobject.TelegramID(12345))
	user.SetTrelloToken("token-abc")
	userRepo.On("FindByTelegramID", mock.Anything, valueobject.TelegramID(12345)).Return(user, nil)

	title := "Updated Title"
	cardMgr.On("UpdateCard", mock.Anything, "token-abc", "card-1", mock.Anything).Return(nil)

	err := uc.Execute(context.Background(), 12345, "card-1", dto.UpdateCardInput{
		Title: &title,
	})

	assert.NoError(t, err)
}

func TestUpdateCard_TrelloNotConnected(t *testing.T) {
	cardMgr := new(MockCardManager)
	userRepo := new(MockUserRepo)
	uc := usecase.NewUpdateCardUseCase(cardMgr, userRepo)

	user := entity.NewUser(valueobject.TelegramID(12345))
	userRepo.On("FindByTelegramID", mock.Anything, valueobject.TelegramID(12345)).Return(user, nil)

	err := uc.Execute(context.Background(), 12345, "card-1", dto.UpdateCardInput{})

	assert.ErrorIs(t, err, domainerror.ErrTrelloNotConnected)
}

func TestUpdateCard_Fails(t *testing.T) {
	cardMgr := new(MockCardManager)
	userRepo := new(MockUserRepo)
	uc := usecase.NewUpdateCardUseCase(cardMgr, userRepo)

	user := entity.NewUser(valueobject.TelegramID(12345))
	user.SetTrelloToken("token-abc")
	userRepo.On("FindByTelegramID", mock.Anything, valueobject.TelegramID(12345)).Return(user, nil)
	cardMgr.On("UpdateCard", mock.Anything, "token-abc", "card-1", mock.Anything).Return(errors.New("api error"))

	err := uc.Execute(context.Background(), 12345, "card-1", dto.UpdateCardInput{})

	assert.Error(t, err)
}
