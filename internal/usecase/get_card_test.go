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

func TestGetCard_HappyPath(t *testing.T) {
	cardMgr := new(MockCardManager)
	userRepo := new(MockUserRepo)
	uc := usecase.NewGetCardUseCase(cardMgr, userRepo)

	user := entity.NewUser(valueobject.TelegramID(12345))
	user.SetTrelloToken("token-abc")
	userRepo.On("FindByTelegramID", mock.Anything, valueobject.TelegramID(12345)).Return(user, nil)

	cardMgr.On("GetCard", mock.Anything, "token-abc", "card-1").Return(&port.CardInfo{
		ID:          "card-1",
		Title:       "Test Card",
		Description: "Description",
		URL:         "https://trello.com/c/1",
		ListID:      "list-1",
		Due:         "2024-12-31",
		Labels:      []string{"Bug"},
		Members:     []string{"john"},
	}, nil)

	output, err := uc.Execute(context.Background(), 12345, "card-1")

	assert.NoError(t, err)
	assert.Equal(t, "card-1", output.ID)
	assert.Equal(t, "Test Card", output.Title)
	assert.Equal(t, "Description", output.Description)
	assert.Equal(t, "2024-12-31", output.Due)
}

func TestGetCard_TrelloNotConnected(t *testing.T) {
	cardMgr := new(MockCardManager)
	userRepo := new(MockUserRepo)
	uc := usecase.NewGetCardUseCase(cardMgr, userRepo)

	user := entity.NewUser(valueobject.TelegramID(12345))
	userRepo.On("FindByTelegramID", mock.Anything, valueobject.TelegramID(12345)).Return(user, nil)

	output, err := uc.Execute(context.Background(), 12345, "card-1")

	assert.ErrorIs(t, err, domainerror.ErrTrelloNotConnected)
	assert.Nil(t, output)
}

func TestGetCard_NotFound(t *testing.T) {
	cardMgr := new(MockCardManager)
	userRepo := new(MockUserRepo)
	uc := usecase.NewGetCardUseCase(cardMgr, userRepo)

	user := entity.NewUser(valueobject.TelegramID(12345))
	user.SetTrelloToken("token-abc")
	userRepo.On("FindByTelegramID", mock.Anything, valueobject.TelegramID(12345)).Return(user, nil)
	cardMgr.On("GetCard", mock.Anything, "token-abc", "card-1").Return(nil, errors.New("not found"))

	output, err := uc.Execute(context.Background(), 12345, "card-1")

	assert.Error(t, err)
	assert.Nil(t, output)
}
