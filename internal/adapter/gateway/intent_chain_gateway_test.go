package gateway

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"telegram-trello-bot/internal/domain/entity"
	"telegram-trello-bot/internal/usecase/dto"
)

type mockIntentParser struct{ mock.Mock }

func (m *mockIntentParser) ParseIntent(ctx context.Context, rawMessage string) (*dto.IntentOutput, error) {
	args := m.Called(ctx, rawMessage)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.IntentOutput), args.Error(1)
}

type mockTaskParser struct{ mock.Mock }

func (m *mockTaskParser) Parse(ctx context.Context, rawMessage string) (*entity.Task, error) {
	args := m.Called(ctx, rawMessage)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Task), args.Error(1)
}

func TestIntentChainGateway_PrimarySuccess(t *testing.T) {
	primary := new(mockIntentParser)
	fallback := new(mockTaskParser)
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	gw := NewIntentChainGateway(primary, fallback, logger)

	expected := &dto.IntentOutput{Action: "move_card", CardName: "test"}
	primary.On("ParseIntent", mock.Anything, "move test to Done").Return(expected, nil)

	result, err := gw.ParseIntent(context.Background(), "move test to Done")

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
	fallback.AssertNotCalled(t, "Parse")
}

func TestIntentChainGateway_FallbackOnPrimaryError(t *testing.T) {
	primary := new(mockIntentParser)
	fallback := new(mockTaskParser)
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	gw := NewIntentChainGateway(primary, fallback, logger)

	primary.On("ParseIntent", mock.Anything, "create a task: do something").
		Return(nil, errors.New("API error"))

	task, _ := entity.NewTask("do something")
	fallback.On("Parse", mock.Anything, "create a task: do something").Return(task, nil)

	result, err := gw.ParseIntent(context.Background(), "create a task: do something")

	assert.NoError(t, err)
	assert.Equal(t, "create_task", result.Action)
	assert.Equal(t, "do something", result.Title)
}

func TestIntentChainGateway_BothFail(t *testing.T) {
	primary := new(mockIntentParser)
	fallback := new(mockTaskParser)
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	gw := NewIntentChainGateway(primary, fallback, logger)

	primary.On("ParseIntent", mock.Anything, "test").
		Return(nil, errors.New("API error"))
	fallback.On("Parse", mock.Anything, "test").
		Return(nil, errors.New("parse error"))

	_, err := gw.ParseIntent(context.Background(), "test")

	assert.Error(t, err)
}
