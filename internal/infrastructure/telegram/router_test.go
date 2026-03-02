package telegram_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"telegram-trello-bot/internal/adapter/controller"
	"telegram-trello-bot/internal/adapter/presenter"
	"telegram-trello-bot/internal/domain/entity"
	"telegram-trello-bot/internal/domain/valueobject"
	"telegram-trello-bot/internal/infrastructure/state"
	infratelegram "telegram-trello-bot/internal/infrastructure/telegram"
	"telegram-trello-bot/internal/usecase"
	"telegram-trello-bot/internal/usecase/port"
)

// -- Mock BotSender --

type mockBotSender struct {
	mock.Mock
	lastMessage tgbotapi.MessageConfig
}

func (m *mockBotSender) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	args := m.Called(c)
	if msg, ok := c.(tgbotapi.MessageConfig); ok {
		m.lastMessage = msg
	}
	return tgbotapi.Message{}, args.Error(0)
}

func (m *mockBotSender) Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error) {
	args := m.Called(c)
	return &tgbotapi.APIResponse{Ok: true}, args.Error(0)
}

// -- Mock dependencies --

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

// -- Setup --

type routerFixture struct {
	router   *infratelegram.Router
	api      *mockBotSender
	parser   *mockParser
	board    *mockBoard
	userRepo *mockUserRepo
	taskLog  *mockTaskLog
}

func setupRouter() *routerFixture {
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
	pres := presenter.NewTelegramPresenter()
	logger := slog.Default()

	router := infratelegram.NewRouter(ctrl, pres, logger)
	api := new(mockBotSender)

	return &routerFixture{
		router: router, api: api,
		parser: parser, board: board,
		userRepo: userRepo, taskLog: taskLog,
	}
}

func makeMessage(userID int64, chatID int64, text string) tgbotapi.Update {
	return tgbotapi.Update{
		Message: &tgbotapi.Message{
			Text: text,
			Chat: &tgbotapi.Chat{ID: chatID},
			From: &tgbotapi.User{ID: userID},
		},
	}
}

func makeCommand(userID int64, chatID int64, cmd string) tgbotapi.Update {
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

func makeCallback(userID int64, chatID int64, data string) tgbotapi.Update {
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

// -- Tests --

func TestRouter_StartCommand(t *testing.T) {
	f := setupRouter()
	f.api.On("Send", mock.Anything).Return(nil)

	f.router.Route(f.api, makeCommand(100, 200, "start"))

	f.api.AssertCalled(t, "Send", mock.MatchedBy(func(c tgbotapi.Chattable) bool {
		msg, ok := c.(tgbotapi.MessageConfig)
		return ok && msg.Text == "Welcome! Send /help to get started."
	}))
}

func TestRouter_HelpCommand(t *testing.T) {
	f := setupRouter()
	f.api.On("Send", mock.Anything).Return(nil)

	f.router.Route(f.api, makeCommand(100, 200, "help"))

	f.api.AssertCalled(t, "Send", mock.MatchedBy(func(c tgbotapi.Chattable) bool {
		msg, ok := c.(tgbotapi.MessageConfig)
		return ok && assert.Contains(t, msg.Text, "Trello Bot")
	}))
}

func TestRouter_BoardsCommand_SendsKeyboard(t *testing.T) {
	f := setupRouter()
	user := entity.NewUser(valueobject.TelegramID(100))
	user.SetTrelloToken("tok")
	f.userRepo.On("FindByTelegramID", mock.Anything, valueobject.TelegramID(100)).Return(user, nil)
	f.board.On("GetBoards", mock.Anything, "tok").Return([]port.BoardInfo{
		{ID: "b1", Name: "Work"},
		{ID: "b2", Name: "Personal"},
	}, nil)
	f.api.On("Send", mock.Anything).Return(nil)

	f.router.Route(f.api, makeCommand(100, 200, "boards"))

	f.api.AssertCalled(t, "Send", mock.MatchedBy(func(c tgbotapi.Chattable) bool {
		msg, ok := c.(tgbotapi.MessageConfig)
		if !ok {
			return false
		}
		kb, ok := msg.ReplyMarkup.(tgbotapi.InlineKeyboardMarkup)
		return ok && len(kb.InlineKeyboard) == 2
	}))
}

func TestRouter_BoardsCommand_Error(t *testing.T) {
	f := setupRouter()
	f.userRepo.On("FindByTelegramID", mock.Anything, mock.Anything).Return(nil, errors.New("db error"))
	f.api.On("Send", mock.Anything).Return(nil)

	f.router.Route(f.api, makeCommand(100, 200, "boards"))

	f.api.AssertCalled(t, "Send", mock.MatchedBy(func(c tgbotapi.Chattable) bool {
		msg, ok := c.(tgbotapi.MessageConfig)
		return ok && assert.Contains(t, msg.Text, "Something went wrong")
	}))
}

func TestRouter_UnknownCommand(t *testing.T) {
	f := setupRouter()
	f.api.On("Send", mock.Anything).Return(nil)

	f.router.Route(f.api, makeCommand(100, 200, "unknown"))

	f.api.AssertCalled(t, "Send", mock.MatchedBy(func(c tgbotapi.Chattable) bool {
		msg, ok := c.(tgbotapi.MessageConfig)
		return ok && msg.Text == "Unknown command. Try /help"
	}))
}

func TestRouter_Message_ShowsPreviewWithKeyboard(t *testing.T) {
	f := setupRouter()
	user := entity.NewUser(valueobject.TelegramID(100))
	user.SetDefaultBoard("b1")
	user.SetDefaultList("l1")
	user.SetTrelloToken("tok")
	f.userRepo.On("FindByTelegramID", mock.Anything, valueobject.TelegramID(100)).Return(user, nil)

	task, _ := entity.NewTask("Buy milk")
	f.parser.On("Parse", mock.Anything, "Buy milk").Return(task, nil)
	f.api.On("Send", mock.Anything).Return(nil)

	f.router.Route(f.api, makeMessage(100, 200, "Buy milk"))

	f.api.AssertCalled(t, "Send", mock.MatchedBy(func(c tgbotapi.Chattable) bool {
		msg, ok := c.(tgbotapi.MessageConfig)
		if !ok {
			return false
		}
		kb, hasKb := msg.ReplyMarkup.(tgbotapi.InlineKeyboardMarkup)
		return assert.Contains(t, msg.Text, "Create this task?") &&
			hasKb && len(kb.InlineKeyboard) == 1 && len(kb.InlineKeyboard[0]) == 3
	}))
}

func TestRouter_Message_Error(t *testing.T) {
	f := setupRouter()
	f.userRepo.On("FindByTelegramID", mock.Anything, mock.Anything).Return(nil, errors.New("not found"))
	f.api.On("Send", mock.Anything).Return(nil)

	f.router.Route(f.api, makeMessage(100, 200, "task"))

	f.api.AssertCalled(t, "Send", mock.MatchedBy(func(c tgbotapi.Chattable) bool {
		msg, ok := c.(tgbotapi.MessageConfig)
		return ok && assert.Contains(t, msg.Text, "Something went wrong")
	}))
}

func TestRouter_CallbackBoardSelect(t *testing.T) {
	f := setupRouter()
	user := entity.NewUser(valueobject.TelegramID(100))
	user.SetTrelloToken("tok")
	f.userRepo.On("FindByTelegramID", mock.Anything, valueobject.TelegramID(100)).Return(user, nil)
	f.userRepo.On("Save", mock.Anything, mock.Anything).Return(nil)
	f.board.On("GetLists", mock.Anything, "tok", "board-1").Return([]port.ListInfo{
		{ID: "l1", Name: "To Do"},
	}, nil)
	f.api.On("Request", mock.Anything).Return(nil)
	f.api.On("Send", mock.Anything).Return(nil)

	f.router.Route(f.api, makeCallback(100, 200, "board:board-1"))

	f.api.AssertCalled(t, "Send", mock.MatchedBy(func(c tgbotapi.Chattable) bool {
		msg, ok := c.(tgbotapi.MessageConfig)
		if !ok {
			return false
		}
		kb, ok := msg.ReplyMarkup.(tgbotapi.InlineKeyboardMarkup)
		return ok && len(kb.InlineKeyboard) == 1 && assert.Contains(t, msg.Text, "select a list")
	}))
}

func TestRouter_CallbackListSelect(t *testing.T) {
	f := setupRouter()
	user := entity.NewUser(valueobject.TelegramID(100))
	f.userRepo.On("FindByTelegramID", mock.Anything, valueobject.TelegramID(100)).Return(user, nil)
	f.userRepo.On("Save", mock.Anything, mock.Anything).Return(nil)
	f.api.On("Request", mock.Anything).Return(nil)
	f.api.On("Send", mock.Anything).Return(nil)

	f.router.Route(f.api, makeCallback(100, 200, "list:list-1"))

	f.api.AssertCalled(t, "Send", mock.MatchedBy(func(c tgbotapi.Chattable) bool {
		msg, ok := c.(tgbotapi.MessageConfig)
		return ok && assert.Contains(t, msg.Text, "all set")
	}))
}

func TestRouter_CallbackConfirmCreate(t *testing.T) {
	f := setupRouter()
	user := entity.NewUser(valueobject.TelegramID(100))
	user.SetDefaultBoard("b1")
	user.SetDefaultList("l1")
	user.SetTrelloToken("tok")
	f.userRepo.On("FindByTelegramID", mock.Anything, valueobject.TelegramID(100)).Return(user, nil)

	// Parse first to store pending task
	task, _ := entity.NewTask("Buy milk")
	f.parser.On("Parse", mock.Anything, "Buy milk").Return(task, nil)
	f.api.On("Send", mock.Anything).Return(nil)
	f.router.Route(f.api, makeMessage(100, 200, "Buy milk"))

	// Now confirm
	f.board.On("CreateCard", mock.Anything, "tok", mock.Anything).
		Return(&port.CardResult{CardID: "c1", CardURL: "https://trello.com/c/x"}, nil)
	f.taskLog.On("Log", mock.Anything, mock.Anything).Return(nil)
	f.api.On("Request", mock.Anything).Return(nil)

	f.router.Route(f.api, makeCallback(100, 200, "confirm:create"))

	// The last Send call should contain the "Task Created!" message
	assert.Contains(t, f.api.lastMessage.Text, "Task Created!")
}

func TestRouter_CallbackConfirmCancel(t *testing.T) {
	f := setupRouter()
	f.api.On("Request", mock.Anything).Return(nil)
	f.api.On("Send", mock.Anything).Return(nil)

	f.router.Route(f.api, makeCallback(100, 200, "confirm:cancel"))

	f.api.AssertCalled(t, "Send", mock.MatchedBy(func(c tgbotapi.Chattable) bool {
		msg, ok := c.(tgbotapi.MessageConfig)
		return ok && msg.Text == "Task cancelled."
	}))
}

func TestRouter_CallbackConfirmEdit(t *testing.T) {
	f := setupRouter()
	f.api.On("Request", mock.Anything).Return(nil)
	f.api.On("Send", mock.Anything).Return(nil)

	f.router.Route(f.api, makeCallback(100, 200, "confirm:edit"))

	f.api.AssertCalled(t, "Send", mock.MatchedBy(func(c tgbotapi.Chattable) bool {
		msg, ok := c.(tgbotapi.MessageConfig)
		return ok && msg.Text == "Send your edited message:"
	}))
}

func TestRouter_NilMessage(t *testing.T) {
	f := setupRouter()
	f.router.Route(f.api, tgbotapi.Update{})
	f.api.AssertNotCalled(t, "Send")
}
