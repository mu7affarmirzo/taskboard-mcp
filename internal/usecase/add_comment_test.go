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

func TestAddComment_HappyPath(t *testing.T) {
	cardMgr := new(MockCardManager)
	userRepo := new(MockUserRepo)
	uc := usecase.NewAddCommentUseCase(cardMgr, userRepo)

	user := entity.NewUser(valueobject.TelegramID(12345))
	user.SetTrelloToken("token-abc")
	userRepo.On("FindByTelegramID", mock.Anything, valueobject.TelegramID(12345)).Return(user, nil)
	cardMgr.On("AddComment", mock.Anything, "token-abc", "card-1", "Great work!").Return(nil)

	err := uc.Execute(context.Background(), 12345, "card-1", "Great work!")

	assert.NoError(t, err)
}

func TestAddComment_TrelloNotConnected(t *testing.T) {
	cardMgr := new(MockCardManager)
	userRepo := new(MockUserRepo)
	uc := usecase.NewAddCommentUseCase(cardMgr, userRepo)

	user := entity.NewUser(valueobject.TelegramID(12345))
	userRepo.On("FindByTelegramID", mock.Anything, valueobject.TelegramID(12345)).Return(user, nil)

	err := uc.Execute(context.Background(), 12345, "card-1", "comment")

	assert.ErrorIs(t, err, domainerror.ErrTrelloNotConnected)
}

func TestAddComment_Fails(t *testing.T) {
	cardMgr := new(MockCardManager)
	userRepo := new(MockUserRepo)
	uc := usecase.NewAddCommentUseCase(cardMgr, userRepo)

	user := entity.NewUser(valueobject.TelegramID(12345))
	user.SetTrelloToken("token-abc")
	userRepo.On("FindByTelegramID", mock.Anything, valueobject.TelegramID(12345)).Return(user, nil)
	cardMgr.On("AddComment", mock.Anything, "token-abc", "card-1", "comment").Return(errors.New("api error"))

	err := uc.Execute(context.Background(), 12345, "card-1", "comment")

	assert.Error(t, err)
}
