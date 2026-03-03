package dto

import "time"

type ConfirmTaskInput struct {
	TelegramID  int64
	Title       string
	Description string
	DueDate     *time.Time
	Priority    string
	Labels      []string
	Members     []string
}
