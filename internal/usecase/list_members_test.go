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

func TestListMembers_HappyPath(t *testing.T) {
	resolver := new(MockMemberResolver)
	userRepo := new(MockUserRepo)
	uc := usecase.NewListMembersUseCase(resolver, userRepo)

	user := entity.NewUser(valueobject.TelegramID(12345))
	user.SetTrelloToken("token-abc")
	userRepo.On("FindByTelegramID", mock.Anything, valueobject.TelegramID(12345)).Return(user, nil)

	resolver.On("GetMembers", mock.Anything, "token-abc", "board-1").Return([]port.MemberInfo{
		{ID: "m1", Username: "john", FullName: "John Doe"},
	}, nil)

	output, err := uc.Execute(context.Background(), 12345, "board-1")

	assert.NoError(t, err)
	assert.Len(t, output.Members, 1)
	assert.Equal(t, "john", output.Members[0].Username)
}

func TestListMembers_TrelloNotConnected(t *testing.T) {
	resolver := new(MockMemberResolver)
	userRepo := new(MockUserRepo)
	uc := usecase.NewListMembersUseCase(resolver, userRepo)

	user := entity.NewUser(valueobject.TelegramID(12345))
	userRepo.On("FindByTelegramID", mock.Anything, valueobject.TelegramID(12345)).Return(user, nil)

	output, err := uc.Execute(context.Background(), 12345, "board-1")

	assert.ErrorIs(t, err, domainerror.ErrTrelloNotConnected)
	assert.Nil(t, output)
}

func TestListMembers_FetchFails(t *testing.T) {
	resolver := new(MockMemberResolver)
	userRepo := new(MockUserRepo)
	uc := usecase.NewListMembersUseCase(resolver, userRepo)

	user := entity.NewUser(valueobject.TelegramID(12345))
	user.SetTrelloToken("token-abc")
	userRepo.On("FindByTelegramID", mock.Anything, valueobject.TelegramID(12345)).Return(user, nil)
	resolver.On("GetMembers", mock.Anything, "token-abc", "board-1").Return(nil, errors.New("api error"))

	output, err := uc.Execute(context.Background(), 12345, "board-1")

	assert.Error(t, err)
	assert.Nil(t, output)
}
