package persistence

import (
	"context"
	"database/sql"

	"telegram-trello-bot/internal/domain/domainerror"
	"telegram-trello-bot/internal/domain/entity"
	"telegram-trello-bot/internal/domain/valueobject"
)

type UserRepoSQLite struct {
	db *sql.DB
}

func NewUserRepoSQLite(db *sql.DB) *UserRepoSQLite {
	return &UserRepoSQLite{db: db}
}

func (r *UserRepoSQLite) FindByTelegramID(ctx context.Context, id valueobject.TelegramID) (*entity.User, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT telegram_id, trello_token, default_board, default_list, use_llm
		 FROM users WHERE telegram_id = ?`, id.Int64())

	var (
		tid   int64
		token string
		board string
		list  string
		llm   bool
	)
	if err := row.Scan(&tid, &token, &board, &list, &llm); err != nil {
		if err == sql.ErrNoRows {
			return nil, domainerror.ErrUserNotFound
		}
		return nil, err
	}

	user := entity.NewUser(valueobject.TelegramID(tid))
	user.SetTrelloToken(token)
	if board != "" {
		user.SetDefaultBoard(board)
	}
	if list != "" {
		user.SetDefaultList(list)
	}
	user.SetUseLLM(llm)
	return user, nil
}

func (r *UserRepoSQLite) Save(ctx context.Context, user *entity.User) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO users (telegram_id, trello_token, default_board, default_list, use_llm, updated_at)
		 VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		 ON CONFLICT(telegram_id) DO UPDATE SET
			trello_token = excluded.trello_token,
			default_board = excluded.default_board,
			default_list = excluded.default_list,
			use_llm = excluded.use_llm,
			updated_at = CURRENT_TIMESTAMP`,
		user.TelegramID().Int64(), user.TrelloToken(),
		user.DefaultBoard(), user.DefaultList(), user.UseLLM(),
	)
	return err
}
