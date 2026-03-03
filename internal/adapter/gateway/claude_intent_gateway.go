package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"telegram-trello-bot/internal/usecase/dto"
)

type ClaudeIntentGateway struct {
	client MessageSender
}

func NewClaudeIntentGateway(client MessageSender) *ClaudeIntentGateway {
	return &ClaudeIntentGateway{client: client}
}

const intentPrompt = `You are a Trello assistant. Analyze the user's message and determine what action they want.

Return ONLY valid JSON with these fields:
{
  "action": "create_task|move_card|update_card|get_card|list_cards|list_lists|list_labels|create_list|archive_card|delete_card|add_comment|assign_card|search_cards|add_label|set_due_date",
  "card_name": "name of the card being referenced (if any)",
  "title": "task title (for create_task only)",
  "description": "description (for create_task/update_card)",
  "due_date": "YYYY-MM-DD (if mentioned)",
  "priority": "low|medium|high (for create_task, default medium)",
  "labels": ["label names"],
  "checklist": ["subtask items"],
  "members": ["usernames without @ symbol"],
  "list_name": "target list name (for move_card, list_cards, create_list)",
  "comment_text": "comment text (for add_comment)",
  "label_name": "single label name (for add_label)",
  "search_query": "search terms (for search_cards)"
}

Only include fields that are relevant to the detected action. Omit null/empty fields.
If the message is clearly about creating a new task or card, use "create_task".
If the message asks about cards in a specific list, use "list_cards".
If the message asks to find or search for cards, use "search_cards".

Today's date is %s. Resolve relative dates like "tomorrow", "next Friday" relative to today.`

type parsedIntentJSON struct {
	Action      string   `json:"action"`
	CardName    string   `json:"card_name"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	DueDate     string   `json:"due_date"`
	Priority    string   `json:"priority"`
	Labels      []string `json:"labels"`
	Checklist   []string `json:"checklist"`
	Members     []string `json:"members"`
	ListName    string   `json:"list_name"`
	CommentText string   `json:"comment_text"`
	LabelName   string   `json:"label_name"`
	SearchQuery string   `json:"search_query"`
}

func (g *ClaudeIntentGateway) ParseIntent(ctx context.Context, rawMessage string) (*dto.IntentOutput, error) {
	prompt := fmt.Sprintf(intentPrompt, time.Now().Format("2006-01-02"))

	response, err := g.client.SendMessage(ctx, prompt, rawMessage)
	if err != nil {
		return nil, fmt.Errorf("claude API: %w", err)
	}

	var parsed parsedIntentJSON
	if err := json.Unmarshal([]byte(response), &parsed); err != nil {
		return nil, fmt.Errorf("parse claude response: %w", err)
	}

	output := &dto.IntentOutput{
		Action:      parsed.Action,
		CardName:    parsed.CardName,
		Title:       parsed.Title,
		Description: parsed.Description,
		Priority:    parsed.Priority,
		Labels:      parsed.Labels,
		Checklist:   parsed.Checklist,
		Members:     parsed.Members,
		ListName:    parsed.ListName,
		CommentText: parsed.CommentText,
		LabelName:   parsed.LabelName,
		SearchQuery: parsed.SearchQuery,
	}

	if parsed.DueDate != "" {
		if due, err := time.Parse("2006-01-02", parsed.DueDate); err == nil {
			output.DueDate = &due
		}
	}

	return output, nil
}
