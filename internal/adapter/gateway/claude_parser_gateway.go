package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"telegram-trello-bot/internal/domain/entity"
	"telegram-trello-bot/internal/domain/valueobject"
)

type MessageSender interface {
	SendMessage(ctx context.Context, systemPrompt, userMessage string) (string, error)
}

type ClaudeParserGateway struct {
	client MessageSender
}

func NewClaudeParserGateway(client MessageSender) *ClaudeParserGateway {
	return &ClaudeParserGateway{client: client}
}

const parsePrompt = `You are a task parser. Extract structured task data from the user's message.
Return ONLY valid JSON with these fields:
{
  "title": "string (required, concise task title)",
  "description": "string (optional, additional details)",
  "due_date": "string (optional, ISO 8601 date like 2025-03-07)",
  "priority": "string (low|medium|high, default medium)",
  "labels": ["string array (optional, extracted tags/categories)"],
  "checklist": ["string array (optional, subtasks if mentioned)"]
}

Today's date is %s. Resolve relative dates like "tomorrow", "next Friday" relative to today.`

type parsedTaskJSON struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	DueDate     string   `json:"due_date"`
	Priority    string   `json:"priority"`
	Labels      []string `json:"labels"`
	Checklist   []string `json:"checklist"`
}

func (g *ClaudeParserGateway) Parse(ctx context.Context, rawMessage string) (*entity.Task, error) {
	prompt := fmt.Sprintf(parsePrompt, time.Now().Format("2006-01-02"))

	response, err := g.client.SendMessage(ctx, prompt, rawMessage)
	if err != nil {
		return nil, fmt.Errorf("claude API: %w", err)
	}

	var parsed parsedTaskJSON
	if err := json.Unmarshal([]byte(response), &parsed); err != nil {
		return nil, fmt.Errorf("parse claude response: %w", err)
	}

	var opts []entity.TaskOption
	if parsed.Description != "" {
		opts = append(opts, entity.WithDescription(parsed.Description))
	}
	if parsed.DueDate != "" {
		if due, err := time.Parse("2006-01-02", parsed.DueDate); err == nil {
			opts = append(opts, entity.WithDueDate(due))
		}
	}
	if parsed.Priority != "" {
		if p, err := valueobject.NewPriority(parsed.Priority); err == nil {
			opts = append(opts, entity.WithPriority(p))
		}
	}
	if len(parsed.Labels) > 0 {
		opts = append(opts, entity.WithLabels(parsed.Labels))
	}
	if len(parsed.Checklist) > 0 {
		opts = append(opts, entity.WithChecklist(parsed.Checklist))
	}

	return entity.NewTask(parsed.Title, opts...)
}
