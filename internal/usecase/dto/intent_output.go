package dto

import "time"

type IntentOutput struct {
	Action      string
	CardName    string
	CardID      string
	Title       string
	Description string
	DueDate     *time.Time
	Priority    string
	Labels      []string
	Checklist   []string
	Members     []string
	ListName    string
	CommentText string
	LabelName   string
	SearchQuery string
}
