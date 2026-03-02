package controller_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"telegram-trello-bot/internal/adapter/controller"
	"telegram-trello-bot/internal/domain/entity"
	"telegram-trello-bot/internal/domain/valueobject"
	"telegram-trello-bot/internal/infrastructure/state"
	"telegram-trello-bot/internal/usecase"
	"telegram-trello-bot/internal/usecase/port"
)

// -- Mocks --

type mockParser struct{ mock.Mock }

func (m *mockParser) Parse(ctx context.Context, msg string) (*entity.Task, error) {
	args := m.Called(ctx, msg)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Task), args.Error(1)
}

type mockBoard struct{ mock.Mock }

func (m *mockBoard) GetBoards(ctx context.Context, token string) ([]port.BoardInfo, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]port.BoardInfo), args.Error(1)
}
func (m *mockBoard) GetLists(ctx context.Context, token, boardID string) ([]port.ListInfo, error) {
	args := m.Called(ctx, token, boardID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]port.ListInfo), args.Error(1)
}
func (m *mockBoard) GetLabels(ctx context.Context, token, boardID string) ([]port.LabelInfo, error) {
	args := m.Called(ctx, token, boardID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]port.LabelInfo), args.Error(1)
}
func (m *mockBoard) MatchLabels(ctx context.Context, token, boardID string, names []string) ([]string, error) {
	args := m.Called(ctx, token, boardID, names)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}
func (m *mockBoard) CreateCard(ctx context.Context, token string, p port.CreateCardParams) (*port.CardResult, error) {
	args := m.Called(ctx, token, p)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*port.CardResult), args.Error(1)
}

type mockUserRepo struct{ mock.Mock }

func (m *mockUserRepo) FindByTelegramID(ctx context.Context, id valueobject.TelegramID) (*entity.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}
func (m *mockUserRepo) Save(ctx context.Context, user *entity.User) error {
	return m.Called(ctx, user).Error(0)
}

type mockTaskLog struct{ mock.Mock }

func (m *mockTaskLog) Log(ctx context.Context, entry port.TaskLogEntry) error {
	return m.Called(ctx, entry).Error(0)
}

// -- Helpers --

func setupController() (*controller.TelegramController, *mockParser, *mockBoard, *mockUserRepo, *mockTaskLog) {
	parser := new(mockParser)
	board := new(mockBoard)
	userRepo := new(mockUserRepo)
	taskLog := new(mockTaskLog)
	pending := state.NewPendingStore()

	createTask := usecase.NewCreateTaskUseCase(parser, board, userRepo, taskLog)
	parseTask := usecase.NewParseTaskUseCase(parser, userRepo)
	confirmTask := usecase.NewConfirmTaskUseCase(board, userRepo, taskLog)
	listBoards := usecase.NewListBoardsUseCase(board, userRepo)
	listLists := usecase.NewListListsUseCase(board, userRepo)
	selectBoard := usecase.NewSelectBoardUseCase(userRepo)
	selectList := usecase.NewSelectListUseCase(userRepo)

	ctrl := controller.NewTelegramController(createTask, parseTask, confirmTask, listBoards, listLists, selectBoard, selectList, pending)
	return ctrl, parser, board, userRepo, taskLog
}

func configuredUser() *entity.User {
	u := entity.NewUser(valueobject.TelegramID(100))
	u.SetDefaultBoard("b1")
	u.SetDefaultList("l1")
	u.SetTrelloToken("tok")
	return u
}

// -- Tests --

func TestHandleMessage_HappyPath(t *testing.T) {
	ctrl, parser, board, userRepo, taskLog := setupController()
	ctx := context.Background()

	userRepo.On("FindByTelegramID", mock.Anything, valueobject.TelegramID(100)).Return(configuredUser(), nil)

	task, _ := entity.NewTask("Do stuff")
	parser.On("Parse", mock.Anything, "Do stuff please").Return(task, nil)

	board.On("CreateCard", mock.Anything, "tok", mock.Anything).
		Return(&port.CardResult{CardID: "c1", CardURL: "url"}, nil)
	taskLog.On("Log", mock.Anything, mock.Anything).Return(nil)

	output, err := ctrl.HandleMessage(ctx, 100, "Do stuff please")

	assert.NoError(t, err)
	assert.Equal(t, "Do stuff", output.TaskTitle)
	assert.Equal(t, "url", output.CardURL)
}

func TestHandleMessage_Error(t *testing.T) {
	ctrl, _, _, userRepo, _ := setupController()
	ctx := context.Background()

	userRepo.On("FindByTelegramID", mock.Anything, mock.Anything).Return(nil, errors.New("db down"))

	_, err := ctrl.HandleMessage(ctx, 100, "task")

	assert.Error(t, err)
}

func TestHandleParseTask_HappyPath(t *testing.T) {
	ctrl, parser, _, userRepo, _ := setupController()
	ctx := context.Background()

	userRepo.On("FindByTelegramID", mock.Anything, valueobject.TelegramID(100)).Return(configuredUser(), nil)

	task, _ := entity.NewTask("Fix bug")
	parser.On("Parse", mock.Anything, "Fix bug please").Return(task, nil)

	output, err := ctrl.HandleParseTask(ctx, 100, "Fix bug please")

	assert.NoError(t, err)
	assert.Equal(t, "Fix bug", output.TaskTitle)
}

func TestHandleConfirmTask_HappyPath(t *testing.T) {
	ctrl, parser, board, userRepo, taskLog := setupController()
	ctx := context.Background()

	user := configuredUser()
	userRepo.On("FindByTelegramID", mock.Anything, valueobject.TelegramID(100)).Return(user, nil)

	// First parse to populate pending store
	task, _ := entity.NewTask("Fix bug")
	parser.On("Parse", mock.Anything, "Fix bug").Return(task, nil)
	_, _ = ctrl.HandleParseTask(ctx, 100, "Fix bug")

	// Now confirm
	board.On("CreateCard", mock.Anything, "tok", mock.Anything).
		Return(&port.CardResult{CardID: "c1", CardURL: "url"}, nil)
	taskLog.On("Log", mock.Anything, mock.Anything).Return(nil)

	output, err := ctrl.HandleConfirmTask(ctx, 100)

	assert.NoError(t, err)
	assert.Equal(t, "Fix bug", output.TaskTitle)
	assert.Equal(t, "url", output.CardURL)
}

func TestHandleConfirmTask_NoPendingTask(t *testing.T) {
	ctrl, _, _, _, _ := setupController()
	ctx := context.Background()

	_, err := ctrl.HandleConfirmTask(ctx, 999)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no pending task")
}

func TestHandleCancelTask_DeletesPending(t *testing.T) {
	ctrl, parser, _, userRepo, _ := setupController()
	ctx := context.Background()

	userRepo.On("FindByTelegramID", mock.Anything, valueobject.TelegramID(100)).Return(configuredUser(), nil)

	task, _ := entity.NewTask("Fix bug")
	parser.On("Parse", mock.Anything, "Fix bug").Return(task, nil)
	_, _ = ctrl.HandleParseTask(ctx, 100, "Fix bug")

	ctrl.HandleCancelTask(100)

	_, err := ctrl.HandleConfirmTask(ctx, 100)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no pending task")
}

func TestHandleListBoards_HappyPath(t *testing.T) {
	ctrl, _, board, userRepo, _ := setupController()
	ctx := context.Background()

	user := entity.NewUser(valueobject.TelegramID(100))
	user.SetTrelloToken("tok")
	userRepo.On("FindByTelegramID", mock.Anything, valueobject.TelegramID(100)).Return(user, nil)
	board.On("GetBoards", mock.Anything, "tok").Return([]port.BoardInfo{
		{ID: "b1", Name: "Work"},
	}, nil)

	output, err := ctrl.HandleListBoards(ctx, 100)

	assert.NoError(t, err)
	assert.Len(t, output.Boards, 1)
	assert.Equal(t, "Work", output.Boards[0].Name)
}

func TestHandleListLists_HappyPath(t *testing.T) {
	ctrl, _, board, userRepo, _ := setupController()
	ctx := context.Background()

	user := entity.NewUser(valueobject.TelegramID(100))
	user.SetTrelloToken("tok")
	userRepo.On("FindByTelegramID", mock.Anything, valueobject.TelegramID(100)).Return(user, nil)
	board.On("GetLists", mock.Anything, "tok", "board-1").Return([]port.ListInfo{
		{ID: "l1", Name: "To Do"},
		{ID: "l2", Name: "Done"},
	}, nil)

	output, err := ctrl.HandleListLists(ctx, 100, "board-1")

	assert.NoError(t, err)
	assert.Len(t, output.Lists, 2)
	assert.Equal(t, "To Do", output.Lists[0].Name)
}

func TestHandleSelectBoard_HappyPath(t *testing.T) {
	ctrl, _, _, userRepo, _ := setupController()
	ctx := context.Background()

	user := entity.NewUser(valueobject.TelegramID(100))
	userRepo.On("FindByTelegramID", mock.Anything, valueobject.TelegramID(100)).Return(user, nil)
	userRepo.On("Save", mock.Anything, mock.Anything).Return(nil)

	err := ctrl.HandleSelectBoard(ctx, 100, "board-x")

	assert.NoError(t, err)
	userRepo.AssertExpectations(t)
}

func TestHandleSelectList_HappyPath(t *testing.T) {
	ctrl, _, _, userRepo, _ := setupController()
	ctx := context.Background()

	user := entity.NewUser(valueobject.TelegramID(100))
	userRepo.On("FindByTelegramID", mock.Anything, valueobject.TelegramID(100)).Return(user, nil)
	userRepo.On("Save", mock.Anything, mock.Anything).Return(nil)

	err := ctrl.HandleSelectList(ctx, 100, "list-y")

	assert.NoError(t, err)
	userRepo.AssertExpectations(t)
}
