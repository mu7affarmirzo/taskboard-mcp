package persistence_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"telegram-trello-bot/internal/domain/domainerror"
	"telegram-trello-bot/internal/domain/entity"
	"telegram-trello-bot/internal/domain/valueobject"
	"telegram-trello-bot/internal/infrastructure/persistence"
	"telegram-trello-bot/internal/usecase/port"
)

func setupTestDB(t *testing.T) (*persistence.UserRepoSQLite, *persistence.TaskLogRepoSQLite) {
	t.Helper()
	db, err := persistence.NewSQLiteDB(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return persistence.NewUserRepoSQLite(db), persistence.NewTaskLogRepoSQLite(db)
}

func TestNewSQLiteDB_InMemory(t *testing.T) {
	db, err := persistence.NewSQLiteDB(":memory:")
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	err = db.Ping()
	assert.NoError(t, err)
}

func TestNewSQLiteDB_MigrationsRun(t *testing.T) {
	db, err := persistence.NewSQLiteDB(":memory:")
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	var name string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='users'").Scan(&name)
	assert.NoError(t, err)
	assert.Equal(t, "users", name)

	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='task_log'").Scan(&name)
	assert.NoError(t, err)
	assert.Equal(t, "task_log", name)
}

func TestUserRepo_Save_And_FindByTelegramID(t *testing.T) {
	userRepo, _ := setupTestDB(t)
	ctx := context.Background()

	user := entity.NewUser(valueobject.NewTelegramID(12345))
	user.SetTrelloToken("tok-abc")
	user.SetDefaultBoard("board-1")
	user.SetDefaultList("list-1")

	err := userRepo.Save(ctx, user)
	require.NoError(t, err)

	found, err := userRepo.FindByTelegramID(ctx, valueobject.NewTelegramID(12345))
	require.NoError(t, err)

	assert.Equal(t, valueobject.NewTelegramID(12345), found.TelegramID())
	assert.Equal(t, "tok-abc", found.TrelloToken())
	assert.Equal(t, "board-1", found.DefaultBoard())
	assert.Equal(t, "list-1", found.DefaultList())
	assert.True(t, found.UseLLM())
}

func TestUserRepo_FindByTelegramID_NotFound(t *testing.T) {
	userRepo, _ := setupTestDB(t)
	ctx := context.Background()

	_, err := userRepo.FindByTelegramID(ctx, valueobject.NewTelegramID(99999))

	assert.ErrorIs(t, err, domainerror.ErrUserNotFound)
}

func TestUserRepo_Save_Upsert(t *testing.T) {
	userRepo, _ := setupTestDB(t)
	ctx := context.Background()

	user := entity.NewUser(valueobject.NewTelegramID(12345))
	user.SetTrelloToken("old-token")
	require.NoError(t, userRepo.Save(ctx, user))

	user.SetTrelloToken("new-token")
	user.SetDefaultBoard("board-2")
	require.NoError(t, userRepo.Save(ctx, user))

	found, err := userRepo.FindByTelegramID(ctx, valueobject.NewTelegramID(12345))
	require.NoError(t, err)

	assert.Equal(t, "new-token", found.TrelloToken())
	assert.Equal(t, "board-2", found.DefaultBoard())
}

func TestUserRepo_Save_UseLLM_Toggle(t *testing.T) {
	userRepo, _ := setupTestDB(t)
	ctx := context.Background()

	user := entity.NewUser(valueobject.NewTelegramID(12345))
	require.NoError(t, userRepo.Save(ctx, user))

	found, _ := userRepo.FindByTelegramID(ctx, valueobject.NewTelegramID(12345))
	assert.True(t, found.UseLLM())

	found.SetUseLLM(false)
	require.NoError(t, userRepo.Save(ctx, found))

	found2, _ := userRepo.FindByTelegramID(ctx, valueobject.NewTelegramID(12345))
	assert.False(t, found2.UseLLM())
}

func TestTaskLogRepo_Log(t *testing.T) {
	userRepo, logRepo := setupTestDB(t)
	ctx := context.Background()

	// Must save user first (foreign key)
	user := entity.NewUser(valueobject.NewTelegramID(12345))
	require.NoError(t, userRepo.Save(ctx, user))

	err := logRepo.Log(ctx, port.TaskLogEntry{
		TelegramID: 12345,
		Message:    "Fix bug",
		CardID:     "card-abc",
	})
	assert.NoError(t, err)
}

func TestTaskLogRepo_Log_MultipleEntries(t *testing.T) {
	userRepo, logRepo := setupTestDB(t)
	ctx := context.Background()

	user := entity.NewUser(valueobject.NewTelegramID(12345))
	require.NoError(t, userRepo.Save(ctx, user))

	for i := 0; i < 3; i++ {
		err := logRepo.Log(ctx, port.TaskLogEntry{
			TelegramID: 12345,
			Message:    "task",
			CardID:     "card",
		})
		require.NoError(t, err)
	}
}

func TestUserRepo_ImplementsPort(t *testing.T) {
	db, err := persistence.NewSQLiteDB(":memory:")
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	var _ port.UserRepository = persistence.NewUserRepoSQLite(db)
	var _ port.TaskLogRepository = persistence.NewTaskLogRepoSQLite(db)
}
