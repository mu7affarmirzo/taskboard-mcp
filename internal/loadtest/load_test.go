package loadtest_test

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/stretchr/testify/mock"
	_ "modernc.org/sqlite"

	"telegram-trello-bot/internal/adapter/controller"
	"telegram-trello-bot/internal/adapter/presenter"
	"telegram-trello-bot/internal/domain/entity"
	"telegram-trello-bot/internal/domain/valueobject"
	"telegram-trello-bot/internal/infrastructure/health"
	"telegram-trello-bot/internal/infrastructure/state"
	infratelegram "telegram-trello-bot/internal/infrastructure/telegram"
	"telegram-trello-bot/internal/usecase"
	"telegram-trello-bot/internal/usecase/dto"
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
func (m *mockBoard) SearchCards(ctx context.Context, token, boardID, query string) ([]port.CardResult, error) {
	args := m.Called(ctx, token, boardID, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]port.CardResult), args.Error(1)
}
func (m *mockBoard) GetCards(ctx context.Context, token, listID string) ([]port.CardResult, error) {
	args := m.Called(ctx, token, listID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]port.CardResult), args.Error(1)
}
func (m *mockBoard) CreateList(ctx context.Context, token, boardID, name string) (*port.ListInfo, error) {
	args := m.Called(ctx, token, boardID, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*port.ListInfo), args.Error(1)
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

// ── No-op BotSender ──────────────────────────────────────────

type noopSender struct {
	sendCount atomic.Int64
}

func (s *noopSender) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	s.sendCount.Add(1)
	return tgbotapi.Message{}, nil
}

func (s *noopSender) Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error) {
	return &tgbotapi.APIResponse{Ok: true}, nil
}

// ── PendingStore Load Tests ──────────────────────────────────

func TestPendingStore_ConcurrentSetGet(t *testing.T) {
	store := state.NewPendingStore()
	const writers = 50
	const readers = 50
	const opsPerGoroutine = 1000

	var wg sync.WaitGroup
	wg.Add(writers + readers)

	// Writers: each writes to their own key range
	for w := 0; w < writers; w++ {
		go func(writerID int) {
			defer wg.Done()
			baseID := int64(writerID * opsPerGoroutine)
			for i := 0; i < opsPerGoroutine; i++ {
				store.Set(baseID+int64(i), state.PendingTask{
					Title: fmt.Sprintf("Task-%d-%d", writerID, i),
				})
			}
		}(w)
	}

	// Readers: read random keys concurrently with writes
	for r := 0; r < readers; r++ {
		go func() {
			defer wg.Done()
			for i := 0; i < opsPerGoroutine; i++ {
				id := int64(rand.Intn(writers * opsPerGoroutine))
				store.Get(id)
			}
		}()
	}

	wg.Wait()

	// Verify writes: spot-check a few keys
	for w := 0; w < writers; w++ {
		key := int64(w*opsPerGoroutine + opsPerGoroutine - 1) // last key per writer
		task, ok := store.Get(key)
		if !ok {
			t.Errorf("expected key %d to exist", key)
			continue
		}
		expected := fmt.Sprintf("Task-%d-%d", w, opsPerGoroutine-1)
		if task.Title != expected {
			t.Errorf("key %d: got title %q, want %q", key, task.Title, expected)
		}
	}
}

func TestPendingStore_ConcurrentSetDelete(t *testing.T) {
	store := state.NewPendingStore()
	const goroutines = 100
	const ops = 500

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for g := 0; g < goroutines; g++ {
		go func(id int64) {
			defer wg.Done()
			for i := 0; i < ops; i++ {
				store.Set(id, state.PendingTask{Title: fmt.Sprintf("T-%d", i)})
				store.Get(id)
				store.Delete(id)
			}
		}(int64(g))
	}

	wg.Wait()

	// After all delete cycles, all keys should be gone
	for g := 0; g < goroutines; g++ {
		if _, ok := store.Get(int64(g)); ok {
			t.Errorf("expected key %d to be deleted", g)
		}
	}
}

func TestPendingStore_HotKey_ConcurrentReadWrite(t *testing.T) {
	// Stress test: many goroutines reading and writing the SAME key
	store := state.NewPendingStore()
	const goroutines = 100
	const opsPerGoroutine = 1000

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for g := 0; g < goroutines; g++ {
		go func(gid int) {
			defer wg.Done()
			for i := 0; i < opsPerGoroutine; i++ {
				if i%2 == 0 {
					store.Set(1, state.PendingTask{Title: fmt.Sprintf("Writer-%d-%d", gid, i)})
				} else {
					store.Get(1)
				}
			}
		}(g)
	}

	wg.Wait()

	// Key 1 should exist with some value
	task, ok := store.Get(1)
	if !ok {
		t.Fatal("expected key 1 to exist")
	}
	if task.Title == "" {
		t.Fatal("expected non-empty title")
	}
}

func BenchmarkPendingStore_Set(b *testing.B) {
	store := state.NewPendingStore()
	task := state.PendingTask{Title: "Benchmark task"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.Set(int64(i), task)
	}
}

func BenchmarkPendingStore_Get(b *testing.B) {
	store := state.NewPendingStore()
	for i := 0; i < 10000; i++ {
		store.Set(int64(i), state.PendingTask{Title: "task"})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.Get(int64(i % 10000))
	}
}

func BenchmarkPendingStore_SetGet_Parallel(b *testing.B) {
	store := state.NewPendingStore()
	task := state.PendingTask{Title: "Benchmark task"}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		id := int64(0)
		for pb.Next() {
			store.Set(id, task)
			store.Get(id)
			id++
		}
	})
}

// ── Router Throughput Tests ──────────────────────────────────

type mockIntentParser struct{ mock.Mock }

func (m *mockIntentParser) ParseIntent(ctx context.Context, rawMessage string) (*dto.IntentOutput, error) {
	args := m.Called(ctx, rawMessage)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.IntentOutput), args.Error(1)
}

type mockCardManager struct{ mock.Mock }

func (m *mockCardManager) GetCard(ctx context.Context, token, cardID string) (*port.CardInfo, error) {
	args := m.Called(ctx, token, cardID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*port.CardInfo), args.Error(1)
}
func (m *mockCardManager) UpdateCard(ctx context.Context, token, cardID string, params port.UpdateCardParams) error {
	return m.Called(ctx, token, cardID, params).Error(0)
}
func (m *mockCardManager) ArchiveCard(ctx context.Context, token, cardID string) error {
	return m.Called(ctx, token, cardID).Error(0)
}
func (m *mockCardManager) DeleteCard(ctx context.Context, token, cardID string) error {
	return m.Called(ctx, token, cardID).Error(0)
}
func (m *mockCardManager) AddComment(ctx context.Context, token, cardID, text string) error {
	return m.Called(ctx, token, cardID, text).Error(0)
}

func newLoadRouter(t *testing.T) (*infratelegram.Router, *noopSender, *mockIntentParser, *mockBoard, *mockUserRepo, *mockTaskLog) {
	t.Helper()

	parser := new(mockParser)
	board := new(mockBoard)
	userRepo := new(mockUserRepo)
	taskLog := new(mockTaskLog)
	intentParser := new(mockIntentParser)
	cardManager := new(mockCardManager)
	pending := state.NewPendingStore()

	memberResolver := new(mockMemberResolver)
	createTask := usecase.NewCreateTaskUseCase(parser, board, memberResolver, userRepo, taskLog)
	parseTask := usecase.NewParseTaskUseCase(parser, userRepo)
	confirmTask := usecase.NewConfirmTaskUseCase(board, memberResolver, userRepo, taskLog)
	listBoards := usecase.NewListBoardsUseCase(board, userRepo)
	listLists := usecase.NewListListsUseCase(board, userRepo)
	selectBoard := usecase.NewSelectBoardUseCase(userRepo)
	selectList := usecase.NewSelectListUseCase(userRepo)
	registerUser := usecase.NewRegisterUserUseCase(userRepo, "test-api-key")
	connectTrello := usecase.NewConnectTrelloUseCase(userRepo)

	parseIntentUC := usecase.NewParseIntentUseCase(intentParser, userRepo)
	executeActionUC := usecase.NewExecuteActionUseCase(board, cardManager, memberResolver, userRepo, taskLog)

	ctrl := controller.NewTelegramController(createTask, parseTask, confirmTask, listBoards, listLists, selectBoard, selectList, registerUser, connectTrello, pending, parseIntentUC, executeActionUC)
	pres := presenter.NewTelegramPresenter()
	logger := slog.Default()
	router := infratelegram.NewRouter(ctrl, pres, logger)
	sender := &noopSender{}

	return router, sender, intentParser, board, userRepo, taskLog
}

func TestRouter_Throughput_Commands(t *testing.T) {
	router, sender, _, _, _, _ := newLoadRouter(t)

	const goroutines = 50
	const commandsPerGoroutine = 200

	var wg sync.WaitGroup
	wg.Add(goroutines)

	start := time.Now()
	for g := 0; g < goroutines; g++ {
		go func(userID int64) {
			defer wg.Done()
			for i := 0; i < commandsPerGoroutine; i++ {
				update := tgbotapi.Update{
					Message: &tgbotapi.Message{
						Chat: &tgbotapi.Chat{ID: userID},
						From: &tgbotapi.User{ID: userID},
						Text: "/help",
						Entities: []tgbotapi.MessageEntity{
							{Type: "bot_command", Offset: 0, Length: 5},
						},
					},
				}
				router.Route(sender, update)
			}
		}(int64(g + 1))
	}

	wg.Wait()
	elapsed := time.Since(start)

	total := int64(goroutines * commandsPerGoroutine)
	rps := float64(total) / elapsed.Seconds()
	t.Logf("Router command throughput: %d requests in %v (%.0f rps)", total, elapsed, rps)

	if sender.sendCount.Load() != total {
		t.Errorf("expected %d sends, got %d", total, sender.sendCount.Load())
	}
}

func TestRouter_Throughput_ParseFlow(t *testing.T) {
	router, sender, parser, _, userRepo, _ := newLoadRouter(t)

	const goroutines = 20
	const flowsPerGoroutine = 100

	// Set up mocks for all users
	parser.On("ParseIntent", mock.Anything, mock.Anything).Return(&dto.IntentOutput{
		Action: "create_task",
		Title:  "Load test task",
	}, nil)

	for g := 0; g < goroutines; g++ {
		userID := int64(g + 1)
		tid := valueobject.NewTelegramID(userID)
		user := entity.NewUser(tid)
		user.SetTrelloToken(fmt.Sprintf("token-%d", userID))
		user.SetDefaultBoard("board1")
		user.SetDefaultList("list1")
		userRepo.On("FindByTelegramID", mock.Anything, tid).Return(user, nil)
	}

	var wg sync.WaitGroup
	wg.Add(goroutines)

	start := time.Now()
	for g := 0; g < goroutines; g++ {
		go func(userID int64) {
			defer wg.Done()
			for i := 0; i < flowsPerGoroutine; i++ {
				update := tgbotapi.Update{
					Message: &tgbotapi.Message{
						Chat: &tgbotapi.Chat{ID: userID},
						From: &tgbotapi.User{ID: userID},
						Text: fmt.Sprintf("Task %d for load test", i),
					},
				}
				router.Route(sender, update)
			}
		}(int64(g + 1))
	}

	wg.Wait()
	elapsed := time.Since(start)

	total := int64(goroutines * flowsPerGoroutine)
	rps := float64(total) / elapsed.Seconds()
	t.Logf("Router parse throughput: %d requests in %v (%.0f rps)", total, elapsed, rps)

	if sender.sendCount.Load() != total {
		t.Errorf("expected %d sends, got %d", total, sender.sendCount.Load())
	}
}

// ── Health Endpoint Load Test ────────────────────────────────

func TestHealth_ConcurrentRequests(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	handler := health.NewHealthHandler(db)
	server := httptest.NewServer(handler)
	defer server.Close()

	const goroutines = 50
	const requestsPerGoroutine = 100

	var wg sync.WaitGroup
	var successCount atomic.Int64
	var failCount atomic.Int64

	wg.Add(goroutines)
	start := time.Now()

	for g := 0; g < goroutines; g++ {
		go func() {
			defer wg.Done()
			client := &http.Client{Timeout: 5 * time.Second}
			for i := 0; i < requestsPerGoroutine; i++ {
				resp, err := client.Get(server.URL)
				if err != nil || resp.StatusCode != http.StatusOK {
					failCount.Add(1)
					if resp != nil {
						resp.Body.Close()
					}
					continue
				}
				resp.Body.Close()
				successCount.Add(1)
			}
		}()
	}

	wg.Wait()
	elapsed := time.Since(start)

	total := int64(goroutines * requestsPerGoroutine)
	rps := float64(total) / elapsed.Seconds()
	t.Logf("Health endpoint: %d requests in %v (%.0f rps), %d success, %d fail",
		total, elapsed, rps, successCount.Load(), failCount.Load())

	if failCount.Load() > 0 {
		t.Errorf("expected 0 failures, got %d", failCount.Load())
	}
}

func BenchmarkHealth_Handler(b *testing.B) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	handler := health.NewHealthHandler(db)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/healthz", nil)
			handler.ServeHTTP(w, r)
			if w.Code != http.StatusOK {
				b.Fatalf("unexpected status: %d", w.Code)
			}
		}
	})
}
