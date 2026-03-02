package persistence

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

func NewSQLiteDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		return nil, fmt.Errorf("set WAL mode: %w", err)
	}
	if err := runMigrations(db); err != nil {
		return nil, fmt.Errorf("migrations: %w", err)
	}
	return db, nil
}

func runMigrations(db *sql.DB) error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS users (
			telegram_id   INTEGER PRIMARY KEY,
			trello_token  TEXT NOT NULL DEFAULT '',
			default_board TEXT NOT NULL DEFAULT '',
			default_list  TEXT NOT NULL DEFAULT '',
			use_llm       BOOLEAN NOT NULL DEFAULT 1,
			created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS task_log (
			id            INTEGER PRIMARY KEY AUTOINCREMENT,
			telegram_id   INTEGER NOT NULL,
			message       TEXT NOT NULL,
			trello_card_id TEXT NOT NULL,
			created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (telegram_id) REFERENCES users(telegram_id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_task_log_telegram_id ON task_log(telegram_id)`,
	}
	for _, m := range migrations {
		if _, err := db.Exec(m); err != nil {
			return fmt.Errorf("migration failed: %w\nSQL: %s", err, m)
		}
	}
	return nil
}
