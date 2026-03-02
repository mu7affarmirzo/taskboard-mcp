package port

import "context"

type TaskLogEntry struct {
	TelegramID int64
	Message    string
	CardID     string
}

type TaskLogRepository interface {
	Log(ctx context.Context, entry TaskLogEntry) error
}
