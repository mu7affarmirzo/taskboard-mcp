package gateway_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"telegram-trello-bot/internal/adapter/gateway"
	"telegram-trello-bot/internal/domain/entity"
)

type mockTaskParser struct{ mock.Mock }

func (m *mockTaskParser) Parse(ctx context.Context, msg string) (*entity.Task, error) {
	args := m.Called(ctx, msg)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Task), args.Error(1)
}

func TestParserChain_PrimarySucceeds(t *testing.T) {
	primary := new(mockTaskParser)
	fallback := new(mockTaskParser)
	logger := slog.Default()

	chain := gateway.NewParserChainGateway(primary, fallback, logger)

	task, _ := entity.NewTask("From primary")
	primary.On("Parse", mock.Anything, "hello").Return(task, nil)

	result, err := chain.Parse(context.Background(), "hello")

	assert.NoError(t, err)
	assert.Equal(t, "From primary", result.Title())
	fallback.AssertNotCalled(t, "Parse")
}

func TestParserChain_PrimaryFails_FallbackSucceeds(t *testing.T) {
	primary := new(mockTaskParser)
	fallback := new(mockTaskParser)
	logger := slog.Default()

	chain := gateway.NewParserChainGateway(primary, fallback, logger)

	primary.On("Parse", mock.Anything, "hello").Return(nil, errors.New("LLM error"))
	task, _ := entity.NewTask("From fallback")
	fallback.On("Parse", mock.Anything, "hello").Return(task, nil)

	result, err := chain.Parse(context.Background(), "hello")

	assert.NoError(t, err)
	assert.Equal(t, "From fallback", result.Title())
	primary.AssertExpectations(t)
	fallback.AssertExpectations(t)
}

func TestParserChain_BothFail(t *testing.T) {
	primary := new(mockTaskParser)
	fallback := new(mockTaskParser)
	logger := slog.Default()

	chain := gateway.NewParserChainGateway(primary, fallback, logger)

	primary.On("Parse", mock.Anything, "").Return(nil, errors.New("LLM error"))
	fallback.On("Parse", mock.Anything, "").Return(nil, errors.New("rule error"))

	_, err := chain.Parse(context.Background(), "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rule error")
}
