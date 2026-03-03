package gateway

import (
	"context"
	"log/slog"

	"telegram-trello-bot/internal/usecase/dto"
	"telegram-trello-bot/internal/usecase/port"
)

type IntentChainGateway struct {
	primary  port.IntentParser
	fallback port.TaskParser
	logger   *slog.Logger
}

func NewIntentChainGateway(primary port.IntentParser, fallback port.TaskParser, logger *slog.Logger) *IntentChainGateway {
	return &IntentChainGateway{primary: primary, fallback: fallback, logger: logger}
}

func (g *IntentChainGateway) ParseIntent(ctx context.Context, rawMessage string) (*dto.IntentOutput, error) {
	intent, err := g.primary.ParseIntent(ctx, rawMessage)
	if err != nil {
		g.logger.Warn("intent parser failed, falling back to task parser", "error", err)
		return g.fallbackToTaskParser(ctx, rawMessage)
	}
	return intent, nil
}

func (g *IntentChainGateway) fallbackToTaskParser(ctx context.Context, rawMessage string) (*dto.IntentOutput, error) {
	task, err := g.fallback.Parse(ctx, rawMessage)
	if err != nil {
		return nil, err
	}

	output := &dto.IntentOutput{
		Action:      "create_task",
		Title:       task.Title(),
		Description: task.Description(),
		DueDate:     task.DueDate(),
		Priority:    string(task.Priority()),
		Labels:      task.Labels(),
		Checklist:   task.Checklist(),
		Members:     task.Members(),
	}

	return output, nil
}
