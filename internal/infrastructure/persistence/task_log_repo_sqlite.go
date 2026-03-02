package persistence

import (
	"context"
	"database/sql"

	"telegram-trello-bot/internal/usecase/port"
)

type TaskLogRepoSQLite struct {
	db *sql.DB
}

func NewTaskLogRepoSQLite(db *sql.DB) *TaskLogRepoSQLite {
	return &TaskLogRepoSQLite{db: db}
}

func (r *TaskLogRepoSQLite) Log(ctx context.Context, entry port.TaskLogEntry) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO task_log (telegram_id, message, trello_card_id) VALUES (?, ?, ?)`,
		entry.TelegramID, entry.Message, entry.CardID,
	)
	return err
}
