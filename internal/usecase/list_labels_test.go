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

func TestListLabels_HappyPath(t *testing.T) {
	board := new(MockBoard)
	userRepo := new(MockUserRepo)
	uc := usecase.NewListLabelsUseCase(board, userRepo)

	user := entity.NewUser(valueobject.TelegramID(12345))
	user.SetTrelloToken("token-abc")
	userRepo.On("FindByTelegramID", mock.Anything, valueobject.TelegramID(12345)).Return(user, nil)

	board.On("GetLabels", mock.Anything, "token-abc", "board-1").Return([]port.LabelInfo{
		{ID: "l1", Name: "Bug", Color: "red"},
		{ID: "l2", Name: "Feature", Color: "green"},
	}, nil)

	output, err := uc.Execute(context.Background(), 12345, "board-1")

	assert.NoError(t, err)
	assert.Len(t, output.Labels, 2)
	assert.Equal(t, "Bug", output.Labels[0].Name)
	assert.Equal(t, "red", output.Labels[0].Color)
}

func TestListLabels_TrelloNotConnected(t *testing.T) {
	board := new(MockBoard)
	userRepo := new(MockUserRepo)
	uc := usecase.NewListLabelsUseCase(board, userRepo)

	user := entity.NewUser(valueobject.TelegramID(12345))
	userRepo.On("FindByTelegramID", mock.Anything, valueobject.TelegramID(12345)).Return(user, nil)

	output, err := uc.Execute(context.Background(), 12345, "board-1")

	assert.ErrorIs(t, err, domainerror.ErrTrelloNotConnected)
	assert.Nil(t, output)
}

func TestListLabels_FetchFails(t *testing.T) {
	board := new(MockBoard)
	userRepo := new(MockUserRepo)
	uc := usecase.NewListLabelsUseCase(board, userRepo)

	user := entity.NewUser(valueobject.TelegramID(12345))
	user.SetTrelloToken("token-abc")
	userRepo.On("FindByTelegramID", mock.Anything, valueobject.TelegramID(12345)).Return(user, nil)
	board.On("GetLabels", mock.Anything, "token-abc", "board-1").Return(nil, errors.New("api error"))

	output, err := uc.Execute(context.Background(), 12345, "board-1")

	assert.Error(t, err)
	assert.Nil(t, output)
}
