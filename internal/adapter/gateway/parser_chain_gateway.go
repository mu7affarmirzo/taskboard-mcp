package gateway

import (
	"context"
	"log/slog"

	"telegram-trello-bot/internal/domain/entity"
	"telegram-trello-bot/internal/usecase/port"
)

type ParserChainGateway struct {
	primary  port.TaskParser
	fallback port.TaskParser
	logger   *slog.Logger
}

func NewParserChainGateway(primary, fallback port.TaskParser, logger *slog.Logger) *ParserChainGateway {
	return &ParserChainGateway{primary: primary, fallback: fallback, logger: logger}
}

func (g *ParserChainGateway) Parse(ctx context.Context, msg string) (*entity.Task, error) {
	task, err := g.primary.Parse(ctx, msg)
	if err != nil {
		g.logger.Warn("primary parser failed, falling back", "error", err)
		return g.fallback.Parse(ctx, msg)
	}
	return task, nil
}
