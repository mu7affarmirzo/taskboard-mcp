package dto

import "time"

type CreateTaskOutput struct {
	CardURL   string
	TaskTitle string
	DueDate   *time.Time
	Priority  string
	Labels    []string
	Members   []string
}
