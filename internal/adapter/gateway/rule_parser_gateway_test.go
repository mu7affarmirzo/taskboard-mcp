package gateway_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"telegram-trello-bot/internal/adapter/gateway"
	"telegram-trello-bot/internal/domain/valueobject"
)

func TestRuleParser_SimpleTitle(t *testing.T) {
	parser := gateway.NewRuleParserGateway()

	task, err := parser.Parse(context.Background(), "Buy groceries")

	require.NoError(t, err)
	assert.Equal(t, "Buy groceries", task.Title())
	assert.Equal(t, valueobject.PriorityMedium, task.Priority())
}

func TestRuleParser_HighPriority_Urgent(t *testing.T) {
	parser := gateway.NewRuleParserGateway()

	task, err := parser.Parse(context.Background(), "Fix server urgent")

	require.NoError(t, err)
	assert.Equal(t, "Fix server", task.Title())
	assert.Equal(t, valueobject.PriorityHigh, task.Priority())
}

func TestRuleParser_HighPriority_Explicit(t *testing.T) {
	parser := gateway.NewRuleParserGateway()

	task, err := parser.Parse(context.Background(), "Deploy app high priority")

	require.NoError(t, err)
	assert.Equal(t, "Deploy app", task.Title())
	assert.Equal(t, valueobject.PriorityHigh, task.Priority())
}

func TestRuleParser_LowPriority(t *testing.T) {
	parser := gateway.NewRuleParserGateway()

	task, err := parser.Parse(context.Background(), "Update docs low priority")

	require.NoError(t, err)
	assert.Equal(t, "Update docs", task.Title())
	assert.Equal(t, valueobject.PriorityLow, task.Priority())
}

func TestRuleParser_Labels(t *testing.T) {
	parser := gateway.NewRuleParserGateway()

	task, err := parser.Parse(context.Background(), "Fix auth #backend #security")

	require.NoError(t, err)
	assert.Equal(t, "Fix auth", task.Title())
	assert.Equal(t, []string{"backend", "security"}, task.Labels())
}

func TestRuleParser_DueDate_Tomorrow(t *testing.T) {
	parser := gateway.NewRuleParserGateway()

	task, err := parser.Parse(context.Background(), "Send report due tomorrow")

	require.NoError(t, err)
	assert.Equal(t, "Send report", task.Title())
	assert.NotNil(t, task.DueDate())
}

func TestRuleParser_DueDate_Weekday(t *testing.T) {
	parser := gateway.NewRuleParserGateway()

	task, err := parser.Parse(context.Background(), "Review PR by friday")

	require.NoError(t, err)
	assert.Equal(t, "Review PR", task.Title())
	assert.NotNil(t, task.DueDate())
}

func TestRuleParser_Combined(t *testing.T) {
	parser := gateway.NewRuleParserGateway()

	task, err := parser.Parse(context.Background(), "Fix payment bug urgent #backend due tomorrow")

	require.NoError(t, err)
	assert.Equal(t, "Fix payment bug", task.Title())
	assert.Equal(t, valueobject.PriorityHigh, task.Priority())
	assert.Equal(t, []string{"backend"}, task.Labels())
	assert.NotNil(t, task.DueDate())
}

func TestRuleParser_Checklist_DashItems(t *testing.T) {
	parser := gateway.NewRuleParserGateway()

	task, err := parser.Parse(context.Background(), "Prepare release\n- Update changelog\n- Tag version\n- Deploy to staging")

	require.NoError(t, err)
	assert.Equal(t, "Prepare release", task.Title())
	assert.Equal(t, []string{"Update changelog", "Tag version", "Deploy to staging"}, task.Checklist())
}

func TestRuleParser_Checklist_AsteriskItems(t *testing.T) {
	parser := gateway.NewRuleParserGateway()

	task, err := parser.Parse(context.Background(), "Shopping list\n* Milk\n* Bread\n* Eggs")

	require.NoError(t, err)
	assert.Equal(t, "Shopping list", task.Title())
	assert.Equal(t, []string{"Milk", "Bread", "Eggs"}, task.Checklist())
}

func TestRuleParser_Checklist_WithPriorityAndLabels(t *testing.T) {
	parser := gateway.NewRuleParserGateway()

	task, err := parser.Parse(context.Background(), "Deploy app urgent #devops\n- Build image\n- Run migrations\n- Verify health")

	require.NoError(t, err)
	assert.Equal(t, "Deploy app", task.Title())
	assert.Equal(t, valueobject.PriorityHigh, task.Priority())
	assert.Equal(t, []string{"devops"}, task.Labels())
	assert.Equal(t, []string{"Build image", "Run migrations", "Verify health"}, task.Checklist())
}

func TestRuleParser_NoChecklist_SingleLine(t *testing.T) {
	parser := gateway.NewRuleParserGateway()

	task, err := parser.Parse(context.Background(), "Just a simple task")

	require.NoError(t, err)
	assert.Nil(t, task.Checklist())
}

func TestRuleParser_EmptyMessage(t *testing.T) {
	parser := gateway.NewRuleParserGateway()

	_, err := parser.Parse(context.Background(), "")

	assert.Error(t, err) // empty title should fail
}
