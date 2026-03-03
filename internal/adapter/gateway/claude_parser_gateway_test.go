package gateway_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"telegram-trello-bot/internal/adapter/gateway"
	"telegram-trello-bot/internal/domain/valueobject"
)

type mockMessageSender struct{ mock.Mock }

func (m *mockMessageSender) SendMessage(ctx context.Context, systemPrompt, userMessage string) (string, error) {
	args := m.Called(ctx, systemPrompt, userMessage)
	return args.String(0), args.Error(1)
}

func TestClaudeParser_HappyPath_AllFields(t *testing.T) {
	sender := new(mockMessageSender)
	parser := gateway.NewClaudeParserGateway(sender)

	response := `{"title":"Fix payment bug","description":"Critical issue in checkout","due_date":"2025-03-15","priority":"high","labels":["backend","urgent"],"checklist":["Reproduce bug","Write fix","Add tests"]}`
	sender.On("SendMessage", mock.Anything, mock.Anything, "Fix the payment bug urgently").Return(response, nil)

	task, err := parser.Parse(context.Background(), "Fix the payment bug urgently")

	require.NoError(t, err)
	assert.Equal(t, "Fix payment bug", task.Title())
	assert.Equal(t, "Critical issue in checkout", task.Description())
	assert.Equal(t, valueobject.PriorityHigh, task.Priority())
	assert.Equal(t, []string{"backend", "urgent"}, task.Labels())
	assert.Equal(t, []string{"Reproduce bug", "Write fix", "Add tests"}, task.Checklist())
	assert.NotNil(t, task.DueDate())
	assert.Equal(t, 2025, task.DueDate().Year())
	assert.Equal(t, 3, int(task.DueDate().Month()))
	assert.Equal(t, 15, task.DueDate().Day())
	sender.AssertExpectations(t)
}

func TestClaudeParser_MinimalFields(t *testing.T) {
	sender := new(mockMessageSender)
	parser := gateway.NewClaudeParserGateway(sender)

	response := `{"title":"Buy groceries"}`
	sender.On("SendMessage", mock.Anything, mock.Anything, "buy groceries").Return(response, nil)

	task, err := parser.Parse(context.Background(), "buy groceries")

	require.NoError(t, err)
	assert.Equal(t, "Buy groceries", task.Title())
	assert.Equal(t, valueobject.PriorityMedium, task.Priority())
	assert.Empty(t, task.Description())
	assert.Nil(t, task.DueDate())
	assert.Nil(t, task.Labels())
	assert.Nil(t, task.Checklist())
}

func TestClaudeParser_APIError(t *testing.T) {
	sender := new(mockMessageSender)
	parser := gateway.NewClaudeParserGateway(sender)

	sender.On("SendMessage", mock.Anything, mock.Anything, "hello").Return("", errors.New("rate limited"))

	_, err := parser.Parse(context.Background(), "hello")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "claude API")
	assert.Contains(t, err.Error(), "rate limited")
}

func TestClaudeParser_InvalidJSON(t *testing.T) {
	sender := new(mockMessageSender)
	parser := gateway.NewClaudeParserGateway(sender)

	sender.On("SendMessage", mock.Anything, mock.Anything, "task").Return("not valid json", nil)

	_, err := parser.Parse(context.Background(), "task")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parse claude response")
}

func TestClaudeParser_EmptyTitle(t *testing.T) {
	sender := new(mockMessageSender)
	parser := gateway.NewClaudeParserGateway(sender)

	response := `{"title":"","priority":"medium"}`
	sender.On("SendMessage", mock.Anything, mock.Anything, "...").Return(response, nil)

	_, err := parser.Parse(context.Background(), "...")

	assert.Error(t, err) // NewTask rejects empty titles
}

func TestClaudeParser_LowPriority(t *testing.T) {
	sender := new(mockMessageSender)
	parser := gateway.NewClaudeParserGateway(sender)

	response := `{"title":"Update docs","priority":"low"}`
	sender.On("SendMessage", mock.Anything, mock.Anything, "update docs").Return(response, nil)

	task, err := parser.Parse(context.Background(), "update docs")

	require.NoError(t, err)
	assert.Equal(t, valueobject.PriorityLow, task.Priority())
}

func TestClaudeParser_InvalidPriority_DefaultsMedium(t *testing.T) {
	sender := new(mockMessageSender)
	parser := gateway.NewClaudeParserGateway(sender)

	response := `{"title":"Some task","priority":"critical"}`
	sender.On("SendMessage", mock.Anything, mock.Anything, "some task").Return(response, nil)

	task, err := parser.Parse(context.Background(), "some task")

	require.NoError(t, err)
	assert.Equal(t, valueobject.PriorityMedium, task.Priority()) // invalid priority ignored, stays default
}

func TestClaudeParser_InvalidDate_Ignored(t *testing.T) {
	sender := new(mockMessageSender)
	parser := gateway.NewClaudeParserGateway(sender)

	response := `{"title":"Task","due_date":"not-a-date"}`
	sender.On("SendMessage", mock.Anything, mock.Anything, "task").Return(response, nil)

	task, err := parser.Parse(context.Background(), "task")

	require.NoError(t, err)
	assert.Equal(t, "Task", task.Title())
	assert.Nil(t, task.DueDate()) // invalid date gracefully ignored
}

func TestClaudeParser_WithMembers(t *testing.T) {
	sender := new(mockMessageSender)
	parser := gateway.NewClaudeParserGateway(sender)

	response := `{"title":"Fix bug","members":["john","jane"]}`
	sender.On("SendMessage", mock.Anything, mock.Anything, "Fix bug @john @jane").Return(response, nil)

	task, err := parser.Parse(context.Background(), "Fix bug @john @jane")

	require.NoError(t, err)
	assert.Equal(t, "Fix bug", task.Title())
	assert.Equal(t, []string{"john", "jane"}, task.Members())
	sender.AssertExpectations(t)
}

func TestClaudeParser_PromptContainsDate(t *testing.T) {
	sender := new(mockMessageSender)
	parser := gateway.NewClaudeParserGateway(sender)

	response := `{"title":"Test"}`
	sender.On("SendMessage", mock.Anything, mock.MatchedBy(func(prompt string) bool {
		// System prompt should contain today's date
		return assert.Contains(t, prompt, "Today's date is")
	}), "test").Return(response, nil)

	_, err := parser.Parse(context.Background(), "test")
	require.NoError(t, err)
	sender.AssertExpectations(t)
}
