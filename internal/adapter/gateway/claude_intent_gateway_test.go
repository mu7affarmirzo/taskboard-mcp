package gateway

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockMessageSender struct{ mock.Mock }

func (m *mockMessageSender) SendMessage(ctx context.Context, system, user string) (string, error) {
	args := m.Called(ctx, system, user)
	return args.String(0), args.Error(1)
}

func TestClaudeIntentGateway_ParseIntent_MoveCard(t *testing.T) {
	sender := new(mockMessageSender)
	gw := NewClaudeIntentGateway(sender)

	sender.On("SendMessage", mock.Anything, mock.Anything, "move login card to Done").
		Return(`{"action":"move_card","card_name":"login card","list_name":"Done"}`, nil)

	result, err := gw.ParseIntent(context.Background(), "move login card to Done")

	assert.NoError(t, err)
	assert.Equal(t, "move_card", result.Action)
	assert.Equal(t, "login card", result.CardName)
	assert.Equal(t, "Done", result.ListName)
}

func TestClaudeIntentGateway_ParseIntent_CreateTask(t *testing.T) {
	sender := new(mockMessageSender)
	gw := NewClaudeIntentGateway(sender)

	sender.On("SendMessage", mock.Anything, mock.Anything, "create task: implement auth").
		Return(`{"action":"create_task","title":"implement auth","priority":"medium"}`, nil)

	result, err := gw.ParseIntent(context.Background(), "create task: implement auth")

	assert.NoError(t, err)
	assert.Equal(t, "create_task", result.Action)
	assert.Equal(t, "implement auth", result.Title)
	assert.Equal(t, "medium", result.Priority)
}

func TestClaudeIntentGateway_ParseIntent_WithDueDate(t *testing.T) {
	sender := new(mockMessageSender)
	gw := NewClaudeIntentGateway(sender)

	sender.On("SendMessage", mock.Anything, mock.Anything, "set due date on auth card to 2026-03-15").
		Return(`{"action":"set_due_date","card_name":"auth card","due_date":"2026-03-15"}`, nil)

	result, err := gw.ParseIntent(context.Background(), "set due date on auth card to 2026-03-15")

	assert.NoError(t, err)
	assert.Equal(t, "set_due_date", result.Action)
	assert.NotNil(t, result.DueDate)
	assert.Equal(t, 2026, result.DueDate.Year())
}

func TestClaudeIntentGateway_ParseIntent_APIError(t *testing.T) {
	sender := new(mockMessageSender)
	gw := NewClaudeIntentGateway(sender)

	sender.On("SendMessage", mock.Anything, mock.Anything, "test").
		Return("", errors.New("network error"))

	_, err := gw.ParseIntent(context.Background(), "test")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "claude API")
}

func TestClaudeIntentGateway_ParseIntent_InvalidJSON(t *testing.T) {
	sender := new(mockMessageSender)
	gw := NewClaudeIntentGateway(sender)

	sender.On("SendMessage", mock.Anything, mock.Anything, "test").
		Return("not json", nil)

	_, err := gw.ParseIntent(context.Background(), "test")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parse claude response")
}

func TestClaudeIntentGateway_ParseIntent_AddComment(t *testing.T) {
	sender := new(mockMessageSender)
	gw := NewClaudeIntentGateway(sender)

	sender.On("SendMessage", mock.Anything, mock.Anything, "add a comment to payment card: API keys updated").
		Return(`{"action":"add_comment","card_name":"payment card","comment_text":"API keys updated"}`, nil)

	result, err := gw.ParseIntent(context.Background(), "add a comment to payment card: API keys updated")

	assert.NoError(t, err)
	assert.Equal(t, "add_comment", result.Action)
	assert.Equal(t, "payment card", result.CardName)
	assert.Equal(t, "API keys updated", result.CommentText)
}
