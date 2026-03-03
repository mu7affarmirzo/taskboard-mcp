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

func TestParseIntent_HappyPath(t *testing.T) {
	intentParser := new(MockIntentParser)
	userRepo := new(MockUserRepo)

	uc := usecase.NewParseIntentUseCase(intentParser, userRepo)

	user := entity.NewUser(valueobject.TelegramID(123))
	user.SetTrelloToken("tok")
	user.SetDefaultBoard("board1")
	user.SetDefaultList("list1")

	userRepo.On("FindByTelegramID", mock.Anything, valueobject.TelegramID(123)).Return(user, nil)
	intentParser.On("ParseIntent", mock.Anything, "move login card to Done").Return(&dto.IntentOutput{
		Action:   "move_card",
		CardName: "login card",
		ListName: "Done",
	}, nil)

	result, err := uc.Execute(context.Background(), dto.IntentInput{
		TelegramID: 123,
		RawMessage: "move login card to Done",
	})

	assert.NoError(t, err)
	assert.Equal(t, "move_card", result.Action)
	assert.Equal(t, "login card", result.CardName)
	assert.Equal(t, "Done", result.ListName)
}

func TestParseIntent_TrelloNotConnected(t *testing.T) {
	intentParser := new(MockIntentParser)
	userRepo := new(MockUserRepo)

	uc := usecase.NewParseIntentUseCase(intentParser, userRepo)

	user := entity.NewUser(valueobject.TelegramID(123))
	userRepo.On("FindByTelegramID", mock.Anything, valueobject.TelegramID(123)).Return(user, nil)

	_, err := uc.Execute(context.Background(), dto.IntentInput{
		TelegramID: 123,
		RawMessage: "test",
	})

	assert.ErrorIs(t, err, domainerror.ErrTrelloNotConnected)
}

func TestParseIntent_UnknownAction(t *testing.T) {
	intentParser := new(MockIntentParser)
	userRepo := new(MockUserRepo)

	uc := usecase.NewParseIntentUseCase(intentParser, userRepo)

	user := entity.NewUser(valueobject.TelegramID(123))
	user.SetTrelloToken("tok")
	user.SetDefaultBoard("board1")
	user.SetDefaultList("list1")

	userRepo.On("FindByTelegramID", mock.Anything, valueobject.TelegramID(123)).Return(user, nil)
	intentParser.On("ParseIntent", mock.Anything, "something weird").Return(&dto.IntentOutput{
		Action: "unknown_action",
	}, nil)

	_, err := uc.Execute(context.Background(), dto.IntentInput{
		TelegramID: 123,
		RawMessage: "something weird",
	})

	assert.ErrorIs(t, err, domainerror.ErrUnknownAction)
}

func TestParseIntent_ParserError(t *testing.T) {
	intentParser := new(MockIntentParser)
	userRepo := new(MockUserRepo)

	uc := usecase.NewParseIntentUseCase(intentParser, userRepo)

	user := entity.NewUser(valueobject.TelegramID(123))
	user.SetTrelloToken("tok")
	user.SetDefaultBoard("board1")
	user.SetDefaultList("list1")

	userRepo.On("FindByTelegramID", mock.Anything, valueobject.TelegramID(123)).Return(user, nil)
	intentParser.On("ParseIntent", mock.Anything, "test").Return(nil, errors.New("API error"))

	_, err := uc.Execute(context.Background(), dto.IntentInput{
		TelegramID: 123,
		RawMessage: "test",
	})

	assert.ErrorIs(t, err, domainerror.ErrParsingFailed)
}
