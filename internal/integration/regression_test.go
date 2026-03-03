package integration_test

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"telegram-trello-bot/internal/adapter/controller"
	"telegram-trello-bot/internal/adapter/presenter"
	"telegram-trello-bot/internal/domain/domainerror"
	"telegram-trello-bot/internal/domain/entity"
	"telegram-trello-bot/internal/domain/valueobject"
	infratelegram "telegram-trello-bot/internal/infrastructure/telegram"
	"telegram-trello-bot/internal/infrastructure/state"
	"telegram-trello-bot/internal/usecase"
	"telegram-trello-bot/internal/usecase/port"
)

// ── Mocks ───────────────────────────────────────────────────

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

type mockMemberResolver struct{ mock.Mock }

func (m *mockMemberResolver) GetMembers(ctx context.Context, token, boardID string) ([]port.MemberInfo, error) {
	args := m.Called(ctx, token, boardID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]port.MemberInfo), args.Error(1)
}
func (m *mockMemberResolver) MatchMembers(ctx context.Context, token, boardID string, names []string) ([]string, error) {
	args := m.Called(ctx, token, boardID, names)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

// ── Recording BotSender ─────────────────────────────────────

type sentMessage struct {
	text     string
	chatID   int64
	hasKB    bool
	kbRows   int
	kbCols   int // columns in first row
}

type recordingSender struct {
	mu       sync.Mutex
	messages []sentMessage
}

func (s *recordingSender) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	msg, ok := c.(tgbotapi.MessageConfig)
	if !ok {
		return tgbotapi.Message{}, nil
	}
	sm := sentMessage{text: msg.Text, chatID: msg.ChatID}
	if kb, ok := msg.ReplyMarkup.(tgbotapi.InlineKeyboardMarkup); ok {
		sm.hasKB = true
		sm.kbRows = len(kb.InlineKeyboard)
		if len(kb.InlineKeyboard) > 0 {
			sm.kbCols = len(kb.InlineKeyboard[0])
		}
	}
	s.mu.Lock()
	s.messages = append(s.messages, sm)
	s.mu.Unlock()
	return tgbotapi.Message{}, nil
}

func (s *recordingSender) Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error) {
	return &tgbotapi.APIResponse{Ok: true}, nil
}

func (s *recordingSender) last() sentMessage {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.messages[len(s.messages)-1]
}

func (s *recordingSender) count() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.messages)
}

func (s *recordingSender) all() []sentMessage {
	s.mu.Lock()
	defer s.mu.Unlock()
	cp := make([]sentMessage, len(s.messages))
	copy(cp, s.messages)
	return cp
}

// ── Fixture ─────────────────────────────────────────────────

type fixture struct {
	router   *infratelegram.Router
	sender   *recordingSender
	parser   *mockParser
	board    *mockBoard
	userRepo *mockUserRepo
	taskLog  *mockTaskLog
}

func newFixture() *fixture {
	parser := new(mockParser)
	board := new(mockBoard)
	memberResolver := new(mockMemberResolver)
	userRepo := new(mockUserRepo)
	taskLog := new(mockTaskLog)
	pending := state.NewPendingStore()

	createTask := usecase.NewCreateTaskUseCase(parser, board, memberResolver, userRepo, taskLog)
	parseTask := usecase.NewParseTaskUseCase(parser, userRepo)
	confirmTask := usecase.NewConfirmTaskUseCase(board, memberResolver, userRepo, taskLog)
	listBoards := usecase.NewListBoardsUseCase(board, userRepo)
	listLists := usecase.NewListListsUseCase(board, userRepo)
	selectBoard := usecase.NewSelectBoardUseCase(userRepo)
	selectList := usecase.NewSelectListUseCase(userRepo)
	registerUser := usecase.NewRegisterUserUseCase(userRepo, "test-api-key")
	connectTrello := usecase.NewConnectTrelloUseCase(userRepo)

	ctrl := controller.NewTelegramController(
		createTask, parseTask, confirmTask,
		listBoards, listLists, selectBoard, selectList,
		registerUser, connectTrello,
		pending,
	)
	pres := presenter.NewTelegramPresenter()
	logger := slog.Default()

	router := infratelegram.NewRouter(ctrl, pres, logger)
	sender := &recordingSender{}

	return &fixture{
		router: router, sender: sender,
		parser: parser, board: board,
		userRepo: userRepo, taskLog: taskLog,
	}
}

func configuredUser(id int64) *entity.User {
	u := entity.NewUser(valueobject.TelegramID(id))
	u.SetDefaultBoard("board-1")
	u.SetDefaultList("list-1")
	u.SetTrelloToken("trello-tok")
	return u
}

func makeMessage(userID, chatID int64, text string) tgbotapi.Update {
	return tgbotapi.Update{
		Message: &tgbotapi.Message{
			Text: text,
			Chat: &tgbotapi.Chat{ID: chatID},
			From: &tgbotapi.User{ID: userID},
		},
	}
}

func makeCommand(userID, chatID int64, cmd string) tgbotapi.Update {
	return tgbotapi.Update{
		Message: &tgbotapi.Message{
			Text: "/" + cmd,
			Chat: &tgbotapi.Chat{ID: chatID},
			From: &tgbotapi.User{ID: userID},
			Entities: []tgbotapi.MessageEntity{
				{Type: "bot_command", Offset: 0, Length: len(cmd) + 1},
			},
		},
	}
}

func makeCallback(userID, chatID int64, data string) tgbotapi.Update {
	return tgbotapi.Update{
		CallbackQuery: &tgbotapi.CallbackQuery{
			ID:   "cb-1",
			From: &tgbotapi.User{ID: userID},
			Data: data,
			Message: &tgbotapi.Message{
				Chat: &tgbotapi.Chat{ID: chatID},
			},
		},
	}
}

// ═══════════════════════════════════════════════════════════
// REGRESSION TESTS — Full Flow
// ═══════════════════════════════════════════════════════════

func TestFullFlow_ParsePreviewConfirmCreate(t *testing.T) {
	f := newFixture()
	user := configuredUser(100)
	f.userRepo.On("FindByTelegramID", mock.Anything, valueobject.TelegramID(100)).Return(user, nil)

	task, _ := entity.NewTask("Buy groceries", entity.WithPriority(valueobject.PriorityHigh))
	f.parser.On("Parse", mock.Anything, "Buy groceries urgent").Return(task, nil)
	f.board.On("MatchLabels", mock.Anything, "trello-tok", "board-1", mock.Anything).Return([]string{}, nil)
	f.board.On("CreateCard", mock.Anything, "trello-tok", mock.Anything).
		Return(&port.CardResult{CardID: "c1", CardURL: "https://trello.com/c/abc"}, nil)
	f.taskLog.On("Log", mock.Anything, mock.Anything).Return(nil)

	// Step 1: Send message → get preview
	f.router.Route(f.sender, makeMessage(100, 200, "Buy groceries urgent"))

	msgs := f.sender.all()
	require.Len(t, msgs, 1)
	assert.Contains(t, msgs[0].text, "Create this task?")
	assert.Contains(t, msgs[0].text, "Buy groceries")
	assert.True(t, msgs[0].hasKB, "preview should have inline keyboard")
	assert.Equal(t, 3, msgs[0].kbCols, "should have Create/Edit/Cancel buttons")

	// Step 2: Confirm creation
	f.router.Route(f.sender, makeCallback(100, 200, "confirm:create"))

	msgs = f.sender.all()
	require.Len(t, msgs, 2)
	assert.Contains(t, msgs[1].text, "Task Created!")
	assert.Contains(t, msgs[1].text, "Buy groceries")
	assert.Contains(t, msgs[1].text, "https://trello.com/c/abc")

	f.board.AssertCalled(t, "CreateCard", mock.Anything, "trello-tok", mock.Anything)
	f.taskLog.AssertCalled(t, "Log", mock.Anything, mock.Anything)
}

func TestFullFlow_ParsePreviewCancel(t *testing.T) {
	f := newFixture()
	user := configuredUser(100)
	f.userRepo.On("FindByTelegramID", mock.Anything, valueobject.TelegramID(100)).Return(user, nil)

	task, _ := entity.NewTask("Cancelled task")
	f.parser.On("Parse", mock.Anything, "Cancelled task").Return(task, nil)

	// Step 1: Send message → preview
	f.router.Route(f.sender, makeMessage(100, 200, "Cancelled task"))
	require.Equal(t, 1, f.sender.count())
	assert.Contains(t, f.sender.last().text, "Create this task?")

	// Step 2: Cancel
	f.router.Route(f.sender, makeCallback(100, 200, "confirm:cancel"))
	require.Equal(t, 2, f.sender.count())
	assert.Equal(t, "Task cancelled.", f.sender.last().text)

	// Step 3: Confirm after cancel should fail gracefully
	f.router.Route(f.sender, makeCallback(100, 200, "confirm:create"))
	require.Equal(t, 3, f.sender.count())
	assert.Contains(t, f.sender.last().text, "Something went wrong")
	assert.Contains(t, f.sender.last().text, "no pending task")

	f.board.AssertNotCalled(t, "CreateCard")
}

func TestFullFlow_ParsePreviewEdit(t *testing.T) {
	f := newFixture()
	user := configuredUser(100)
	f.userRepo.On("FindByTelegramID", mock.Anything, valueobject.TelegramID(100)).Return(user, nil)

	task1, _ := entity.NewTask("Original task")
	f.parser.On("Parse", mock.Anything, "Original task").Return(task1, nil)

	task2, _ := entity.NewTask("Edited task")
	f.parser.On("Parse", mock.Anything, "Edited task").Return(task2, nil)

	f.board.On("MatchLabels", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]string{}, nil)
	f.board.On("CreateCard", mock.Anything, "trello-tok", mock.Anything).
		Return(&port.CardResult{CardID: "c2", CardURL: "https://trello.com/c/edited"}, nil)
	f.taskLog.On("Log", mock.Anything, mock.Anything).Return(nil)

	// Step 1: Parse original
	f.router.Route(f.sender, makeMessage(100, 200, "Original task"))
	assert.Contains(t, f.sender.last().text, "Original task")

	// Step 2: Choose edit
	f.router.Route(f.sender, makeCallback(100, 200, "confirm:edit"))
	assert.Equal(t, "Send your edited message:", f.sender.last().text)

	// Step 3: Send edited message → new preview
	f.router.Route(f.sender, makeMessage(100, 200, "Edited task"))
	assert.Contains(t, f.sender.last().text, "Edited task")
	assert.Contains(t, f.sender.last().text, "Create this task?")

	// Step 4: Confirm the edited version
	f.router.Route(f.sender, makeCallback(100, 200, "confirm:create"))
	assert.Contains(t, f.sender.last().text, "Task Created!")
	assert.Contains(t, f.sender.last().text, "Edited task")
}

func TestFullFlow_BoardSelection_ThenListSelection(t *testing.T) {
	f := newFixture()
	user := entity.NewUser(valueobject.TelegramID(100))
	user.SetTrelloToken("tok")
	f.userRepo.On("FindByTelegramID", mock.Anything, valueobject.TelegramID(100)).Return(user, nil)
	f.userRepo.On("Save", mock.Anything, mock.Anything).Return(nil)

	f.board.On("GetBoards", mock.Anything, "tok").Return([]port.BoardInfo{
		{ID: "b1", Name: "Work"},
		{ID: "b2", Name: "Personal"},
		{ID: "b3", Name: "Side Project"},
	}, nil)
	f.board.On("GetLists", mock.Anything, "tok", "b2").Return([]port.ListInfo{
		{ID: "l1", Name: "To Do"},
		{ID: "l2", Name: "In Progress"},
		{ID: "l3", Name: "Done"},
	}, nil)

	// Step 1: /boards command
	f.router.Route(f.sender, makeCommand(100, 200, "boards"))
	msgs := f.sender.all()
	require.Len(t, msgs, 1)
	assert.Contains(t, msgs[0].text, "Your Trello Boards")
	assert.Contains(t, msgs[0].text, "Work")
	assert.Contains(t, msgs[0].text, "Personal")
	assert.Contains(t, msgs[0].text, "Side Project")
	assert.True(t, msgs[0].hasKB)
	assert.Equal(t, 3, msgs[0].kbRows, "3 boards = 3 rows")

	// Step 2: Select a board
	f.router.Route(f.sender, makeCallback(100, 200, "board:b2"))
	msgs = f.sender.all()
	require.Len(t, msgs, 2)
	assert.Contains(t, msgs[1].text, "select a list")
	assert.True(t, msgs[1].hasKB)
	assert.Equal(t, 3, msgs[1].kbRows, "3 lists = 3 rows")

	// Step 3: Select a list
	f.router.Route(f.sender, makeCallback(100, 200, "list:l2"))
	msgs = f.sender.all()
	require.Len(t, msgs, 3)
	assert.Contains(t, msgs[2].text, "all set")

	f.userRepo.AssertNumberOfCalls(t, "Save", 2) // board save + list save
}

// ═══════════════════════════════════════════════════════════
// REGRESSION TESTS — Error & Edge Cases
// ═══════════════════════════════════════════════════════════

func TestRegression_UserNotFound_MessageFlow(t *testing.T) {
	f := newFixture()
	f.userRepo.On("FindByTelegramID", mock.Anything, mock.Anything).
		Return(nil, errors.New("user not found"))

	f.router.Route(f.sender, makeMessage(100, 200, "Create task"))

	require.Equal(t, 1, f.sender.count())
	assert.Contains(t, f.sender.last().text, "Something went wrong")
	f.parser.AssertNotCalled(t, "Parse")
}

func TestRegression_UserTrelloNotConnected(t *testing.T) {
	f := newFixture()
	user := entity.NewUser(valueobject.TelegramID(100))
	// No token set
	f.userRepo.On("FindByTelegramID", mock.Anything, valueobject.TelegramID(100)).Return(user, nil)

	f.router.Route(f.sender, makeMessage(100, 200, "Task without token"))

	require.Equal(t, 1, f.sender.count())
	assert.Contains(t, f.sender.last().text, "Something went wrong")
	assert.Contains(t, f.sender.last().text, "trello account not connected")
}

func TestRegression_UserNoBoardConfigured(t *testing.T) {
	f := newFixture()
	user := entity.NewUser(valueobject.TelegramID(100))
	user.SetTrelloToken("tok")
	// No board or list set
	f.userRepo.On("FindByTelegramID", mock.Anything, valueobject.TelegramID(100)).Return(user, nil)

	f.router.Route(f.sender, makeMessage(100, 200, "Task without board"))

	require.Equal(t, 1, f.sender.count())
	assert.Contains(t, f.sender.last().text, "Something went wrong")
	assert.Contains(t, f.sender.last().text, "board not configured")
}

func TestRegression_ParserFailure(t *testing.T) {
	f := newFixture()
	user := configuredUser(100)
	f.userRepo.On("FindByTelegramID", mock.Anything, valueobject.TelegramID(100)).Return(user, nil)
	f.parser.On("Parse", mock.Anything, "???").Return(nil, errors.New("cannot parse"))

	f.router.Route(f.sender, makeMessage(100, 200, "???"))

	require.Equal(t, 1, f.sender.count())
	assert.Contains(t, f.sender.last().text, "Something went wrong")
	f.board.AssertNotCalled(t, "CreateCard")
}

func TestRegression_TrelloAPIFailure_OnConfirm(t *testing.T) {
	f := newFixture()
	user := configuredUser(100)
	f.userRepo.On("FindByTelegramID", mock.Anything, valueobject.TelegramID(100)).Return(user, nil)

	task, _ := entity.NewTask("Important task")
	f.parser.On("Parse", mock.Anything, "Important task").Return(task, nil)
	f.board.On("MatchLabels", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]string{}, nil)
	f.board.On("CreateCard", mock.Anything, mock.Anything, mock.Anything).
		Return(nil, errors.New("trello 500: internal server error"))

	// Parse succeeds
	f.router.Route(f.sender, makeMessage(100, 200, "Important task"))
	assert.Contains(t, f.sender.last().text, "Create this task?")

	// Confirm fails because Trello is down
	f.router.Route(f.sender, makeCallback(100, 200, "confirm:create"))
	assert.Contains(t, f.sender.last().text, "Something went wrong")
	assert.Contains(t, f.sender.last().text, "failed to create card")
}

func TestRegression_ConfirmWithNoPendingTask(t *testing.T) {
	f := newFixture()
	// No parse happened — directly send confirm:create
	f.router.Route(f.sender, makeCallback(100, 200, "confirm:create"))

	require.Equal(t, 1, f.sender.count())
	assert.Contains(t, f.sender.last().text, "Something went wrong")
	assert.Contains(t, f.sender.last().text, "no pending task")
}

func TestRegression_DoubleConfirmSameTask(t *testing.T) {
	f := newFixture()
	user := configuredUser(100)
	f.userRepo.On("FindByTelegramID", mock.Anything, valueobject.TelegramID(100)).Return(user, nil)

	task, _ := entity.NewTask("One-time task")
	f.parser.On("Parse", mock.Anything, "One-time task").Return(task, nil)
	f.board.On("MatchLabels", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]string{}, nil)
	f.board.On("CreateCard", mock.Anything, "trello-tok", mock.Anything).
		Return(&port.CardResult{CardID: "c1", CardURL: "url"}, nil)
	f.taskLog.On("Log", mock.Anything, mock.Anything).Return(nil)

	// Parse → confirm → confirm (second should fail)
	f.router.Route(f.sender, makeMessage(100, 200, "One-time task"))
	f.router.Route(f.sender, makeCallback(100, 200, "confirm:create"))
	assert.Contains(t, f.sender.last().text, "Task Created!")

	f.router.Route(f.sender, makeCallback(100, 200, "confirm:create"))
	assert.Contains(t, f.sender.last().text, "Something went wrong")
	assert.Contains(t, f.sender.last().text, "no pending task")

	// CreateCard should only be called once
	f.board.AssertNumberOfCalls(t, "CreateCard", 1)
}

func TestRegression_UnknownCommand(t *testing.T) {
	f := newFixture()
	f.router.Route(f.sender, makeCommand(100, 200, "foobar"))

	require.Equal(t, 1, f.sender.count())
	assert.Equal(t, "Unknown command. Try /help", f.sender.last().text)
}

func TestRegression_EmptyUpdate_NoResponse(t *testing.T) {
	f := newFixture()
	f.router.Route(f.sender, tgbotapi.Update{})
	assert.Equal(t, 0, f.sender.count())
}

func TestRegression_BoardsCommand_UserNotFound(t *testing.T) {
	f := newFixture()
	f.userRepo.On("FindByTelegramID", mock.Anything, mock.Anything).
		Return(nil, errors.New("db error"))

	f.router.Route(f.sender, makeCommand(100, 200, "boards"))

	require.Equal(t, 1, f.sender.count())
	assert.Contains(t, f.sender.last().text, "Something went wrong")
}

func TestRegression_HelpCommand_ReturnsAllCommands(t *testing.T) {
	f := newFixture()
	f.router.Route(f.sender, makeCommand(100, 200, "help"))

	require.Equal(t, 1, f.sender.count())
	text := f.sender.last().text
	assert.Contains(t, text, "Trello Bot")
	assert.Contains(t, text, "/start")
	assert.Contains(t, text, "/connect")
	assert.Contains(t, text, "/boards")
	assert.Contains(t, text, "/help")
}

func TestRegression_StartCommand_NewUser(t *testing.T) {
	f := newFixture()
	f.userRepo.On("FindByTelegramID", mock.Anything, valueobject.TelegramID(100)).
		Return(nil, domainerror.ErrUserNotFound)
	f.userRepo.On("Save", mock.Anything, mock.Anything).Return(nil)

	f.router.Route(f.sender, makeCommand(100, 200, "start"))

	require.Equal(t, 1, f.sender.count())
	assert.Contains(t, f.sender.last().text, "Welcome")
	assert.Contains(t, f.sender.last().text, "authorize")
}

func TestRegression_TaskWithLabelsAndDueDate(t *testing.T) {
	f := newFixture()
	user := configuredUser(100)
	f.userRepo.On("FindByTelegramID", mock.Anything, valueobject.TelegramID(100)).Return(user, nil)

	due := time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)
	task, _ := entity.NewTask("Deploy v2",
		entity.WithPriority(valueobject.PriorityHigh),
		entity.WithDueDate(due),
		entity.WithLabels([]string{"backend", "urgent"}),
	)
	f.parser.On("Parse", mock.Anything, "Deploy v2 by June 15 #backend #urgent").Return(task, nil)
	f.board.On("MatchLabels", mock.Anything, "trello-tok", "board-1", []string{"backend", "urgent"}).
		Return([]string{"lbl-1", "lbl-2"}, nil)
	f.board.On("CreateCard", mock.Anything, "trello-tok", mock.MatchedBy(func(p port.CreateCardParams) bool {
		return p.Title == "Deploy v2" &&
			p.Position == "top" && // high priority → top
			len(p.LabelIDs) == 2 &&
			p.DueDate != nil
	})).Return(&port.CardResult{CardID: "c1", CardURL: "url"}, nil)
	f.taskLog.On("Log", mock.Anything, mock.Anything).Return(nil)

	// Parse
	f.router.Route(f.sender, makeMessage(100, 200, "Deploy v2 by June 15 #backend #urgent"))
	preview := f.sender.last().text
	assert.Contains(t, preview, "Deploy v2")
	assert.Contains(t, preview, "high")
	assert.Contains(t, preview, "Jun 15, 2025")
	assert.Contains(t, preview, "backend, urgent")

	// Confirm
	f.router.Route(f.sender, makeCallback(100, 200, "confirm:create"))
	assert.Contains(t, f.sender.last().text, "Task Created!")
}

func TestRegression_TaskWithChecklist(t *testing.T) {
	f := newFixture()
	user := configuredUser(100)
	f.userRepo.On("FindByTelegramID", mock.Anything, valueobject.TelegramID(100)).Return(user, nil)

	task, _ := entity.NewTask("Release prep",
		entity.WithChecklist([]string{"run tests", "update changelog", "tag release"}),
	)
	f.parser.On("Parse", mock.Anything, "Release prep with steps").Return(task, nil)

	f.router.Route(f.sender, makeMessage(100, 200, "Release prep with steps"))
	preview := f.sender.last().text
	assert.Contains(t, preview, "Checklist:")
	assert.Contains(t, preview, "run tests")
	assert.Contains(t, preview, "update changelog")
	assert.Contains(t, preview, "tag release")
}

// ═══════════════════════════════════════════════════════════
// REGRESSION TESTS — Concurrent Users
// ═══════════════════════════════════════════════════════════

func TestRegression_ConcurrentUsers_IndependentFlows(t *testing.T) {
	f := newFixture()

	// Set up 10 concurrent users
	const numUsers = 10
	for i := int64(1); i <= numUsers; i++ {
		user := configuredUser(i)
		f.userRepo.On("FindByTelegramID", mock.Anything, valueobject.TelegramID(i)).Return(user, nil)
	}
	task, _ := entity.NewTask("Task from user")
	f.parser.On("Parse", mock.Anything, "Task from user").Return(task, nil)
	f.board.On("MatchLabels", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]string{}, nil)
	f.board.On("CreateCard", mock.Anything, mock.Anything, mock.Anything).
		Return(&port.CardResult{CardID: "c1", CardURL: "url"}, nil)
	f.taskLog.On("Log", mock.Anything, mock.Anything).Return(nil)

	var wg sync.WaitGroup
	for i := int64(1); i <= numUsers; i++ {
		wg.Add(1)
		go func(userID int64) {
			defer wg.Done()
			// Each user: parse → confirm
			f.router.Route(f.sender, makeMessage(userID, userID*10, "Task from user"))
			f.router.Route(f.sender, makeCallback(userID, userID*10, "confirm:create"))
		}(i)
	}
	wg.Wait()

	// Each user should have gotten 2 messages (preview + created)
	assert.Equal(t, numUsers*2, f.sender.count())

	// Verify at least some "Task Created!" messages exist
	created := 0
	for _, m := range f.sender.all() {
		if strings.Contains(m.text, "Task Created!") {
			created++
		}
	}
	assert.Equal(t, numUsers, created, "each user should have exactly one 'Task Created!' message")
}

func TestRegression_ConcurrentUsers_ParseAndCancelMixed(t *testing.T) {
	f := newFixture()

	for i := int64(1); i <= 5; i++ {
		user := configuredUser(i)
		f.userRepo.On("FindByTelegramID", mock.Anything, valueobject.TelegramID(i)).Return(user, nil)
	}
	task, _ := entity.NewTask("Task")
	f.parser.On("Parse", mock.Anything, "Task").Return(task, nil)
	f.board.On("MatchLabels", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]string{}, nil)
	f.board.On("CreateCard", mock.Anything, mock.Anything, mock.Anything).
		Return(&port.CardResult{CardID: "c1", CardURL: "url"}, nil)
	f.taskLog.On("Log", mock.Anything, mock.Anything).Return(nil)

	var wg sync.WaitGroup
	for i := int64(1); i <= 5; i++ {
		wg.Add(1)
		go func(userID int64) {
			defer wg.Done()
			f.router.Route(f.sender, makeMessage(userID, userID*10, "Task"))
			if userID%2 == 0 {
				f.router.Route(f.sender, makeCallback(userID, userID*10, "confirm:cancel"))
			} else {
				f.router.Route(f.sender, makeCallback(userID, userID*10, "confirm:create"))
			}
		}(i)
	}
	wg.Wait()

	msgs := f.sender.all()
	assert.Equal(t, 10, len(msgs)) // 5 previews + 5 responses

	created := 0
	cancelled := 0
	for _, m := range msgs {
		if strings.Contains(m.text, "Task Created!") {
			created++
		}
		if m.text == "Task cancelled." {
			cancelled++
		}
	}
	assert.Equal(t, 3, created, "users 1,3,5 should create")
	assert.Equal(t, 2, cancelled, "users 2,4 should cancel")
}
