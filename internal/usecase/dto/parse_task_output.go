package dto

import "time"

type ParseTaskOutput struct {
	TaskTitle   string
	Description string
	DueDate     *time.Time
	Priority    string
	Labels      []string
	Checklist   []string
	Members     []string
	BoardID     string
	ListID      string
}
