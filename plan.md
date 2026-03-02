Here's the full plan — copy away:

---

# Telegram → Trello Task Bot (Go) — Clean Architecture

## 1. Project Overview

A Telegram bot built in Go following **Clean Architecture** (Uncle Bob) principles. The system receives natural language messages via Telegram, parses them into structured task data, and creates Trello cards — all while keeping business logic completely independent of frameworks, APIs, and infrastructure.

### Clean Architecture Principles Applied

- **Independence of Frameworks**: Business logic doesn't depend on Telegram SDK, Trello API, or any external library
- **Testability**: Use cases can be tested without UI, database, or external services
- **Independence of UI**: Telegram can be swapped for Slack, CLI, or REST API without touching business logic
- **Independence of Database**: SQLite can be swapped for PostgreSQL or Redis without affecting upper layers
- **Dependency Rule**: Dependencies point inward — outer layers depend on inner layers, never the reverse

### Layers (Inside → Out)

```
┌─────────────────────────────────────────────────┐
│                   Domain Layer                   │
│          (Entities, Value Objects, Errors)        │
├─────────────────────────────────────────────────┤
│                 Use Case Layer                   │
│     (Application Business Rules, Ports/Interfaces)│
├─────────────────────────────────────────────────┤
│                 Adapter Layer                    │
│   (Controllers, Presenters, Gateways, Repos)     │
├─────────────────────────────────────────────────┤
│              Infrastructure Layer                │
│    (Telegram SDK, Trello HTTP, SQLite, Claude)    │
└─────────────────────────────────────────────────┘
```

---

## 2. Architecture Diagram

```
                    ┌─────────────────┐
                    │    Telegram      │
                    │   (Delivery)     │
                    └────────┬────────┘
                             │
                    ┌────────▼────────┐
                    │   Controller     │
                    │ (Adapter Layer)  │
                    └────────┬────────┘
                             │ calls
                    ┌────────▼────────┐
                    │    Use Cases     │  ◄── Pure business logic
                    │  (App Layer)     │      No external dependencies
                    └───┬────┬────┬───┘
                        │    │    │
              ┌─────────┘    │    └─────────┐
              │              │              │
     ┌────────▼───┐  ┌──────▼─────┐  ┌─────▼──────┐
     │ TaskParser  │  │ TaskBoard  │  │ UserStore   │
     │  (Port)     │  │  (Port)    │  │  (Port)     │
     └────────┬───┘  └──────┬─────┘  └─────┬──────┘
              │              │              │
     ┌────────▼───┐  ┌──────▼─────┐  ┌─────▼──────┐
     │ Claude/Rule │  │ Trello HTTP│  │  SQLite     │
     │ (Infra)     │  │ (Infra)    │  │  (Infra)    │
     └────────────┘  └────────────┘  └────────────┘
```

---

## 3. Tech Stack

| Component | Technology |
|---|---|
| Language | Go 1.22+ |
| Telegram SDK | `github.com/go-telegram-bot-api/telegram-bot-api/v5` |
| HTTP Client | `net/http` (stdlib) |
| Trello API | Direct REST via `net/http` |
| LLM Parsing (opt) | Claude API or rule-based |
| Config | `github.com/spf13/viper` + `.env` |
| Database | SQLite via `modernc.org/sqlite` |
| DI | Constructor injection (no framework) |
| Logging | `log/slog` (stdlib) |
| Testing | `testing` + `testify` + `mockery` |
| Containerization | Docker multi-stage build |
| CI/CD | GitHub Actions |

---

## 4. Project Structure

```
telegram-trello-bot/
│
├── cmd/
│   └── bot/
│       └── main.go                         # Composition root — wires everything
│
├── internal/
│   │
│   ├── domain/                             # ━━━ DOMAIN LAYER ━━━
│   │   ├── entity/
│   │   │   ├── task.go                     # Task entity (core business object)
│   │   │   ├── user.go                     # User entity
│   │   │   ├── board.go                    # Board entity
│   │   │   └── label.go                    # Label value object
│   │   ├── valueobject/
│   │   │   ├── priority.go                 # Priority enum (Low/Medium/High)
│   │   │   ├── task_id.go                  # Typed ID
│   │   │   └── telegram_id.go             # Typed Telegram user ID
│   │   └── domainerror/
│   │       └── errors.go                   # Domain-specific errors
│   │
│   ├── usecase/                            # ━━━ USE CASE LAYER ━━━
│   │   ├── port/                           # Ports (interfaces) — driven side
│   │   │   ├── task_parser.go              # Port: parse message → Task
│   │   │   ├── task_board.go               # Port: create/manage cards on board
│   │   │   ├── user_repository.go          # Port: persist user preferences
│   │   │   └── task_log_repository.go      # Port: persist task creation log
│   │   ├── create_task.go                  # UseCase: parse message + create card
│   │   ├── create_task_test.go             # Unit test with mocked ports
│   │   ├── select_board.go                 # UseCase: user selects default board
│   │   ├── select_list.go                  # UseCase: user selects default list
│   │   ├── list_boards.go                  # UseCase: fetch available boards
│   │   └── dto/
│   │       ├── create_task_input.go        # Input DTO
│   │       └── create_task_output.go       # Output DTO
│   │
│   ├── adapter/                            # ━━━ ADAPTER LAYER ━━━
│   │   ├── controller/
│   │   │   └── telegram_controller.go      # Maps Telegram updates → use case calls
│   │   ├── presenter/
│   │   │   └── telegram_presenter.go       # Formats use case output → Telegram messages
│   │   └── gateway/
│   │       ├── trello_gateway.go           # Implements TaskBoard port via Trello API
│   │       ├── claude_parser_gateway.go    # Implements TaskParser port via Claude
│   │       ├── rule_parser_gateway.go      # Implements TaskParser port via regex
│   │       └── parser_chain_gateway.go     # Chain: try LLM first, fallback to rules
│   │
│   └── infrastructure/                     # ━━━ INFRASTRUCTURE LAYER ━━━
│       ├── telegram/
│       │   ├── bot.go                      # Telegram bot init, polling/webhook
│       │   ├── router.go                   # Routes updates to controller methods
│       │   └── keyboard.go                 # Inline keyboard builders
│       ├── trello/
│       │   ├── client.go                   # Low-level Trello HTTP client
│       │   └── models.go                   # Trello API response structs
│       ├── claude/
│       │   └── client.go                   # Low-level Claude HTTP client
│       ├── persistence/
│       │   ├── sqlite.go                   # SQLite connection + migrations
│       │   ├── user_repo_sqlite.go         # Implements UserRepository port
│       │   └── task_log_repo_sqlite.go     # Implements TaskLogRepository port
│       └── config/
│           └── config.go                   # Env/config loading with viper
│
├── pkg/                                    # Shared utilities (no business logic)
│   ├── httputil/
│   │   └── client.go                       # Reusable HTTP client with retry
│   └── timeutil/
│       └── parse.go                        # Natural language date parsing
│
├── deployments/
│   ├── Dockerfile
│   └── docker-compose.yml
│
├── .env.example
├── .gitignore
├── .golangci.yml                           # Linter config
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

---

## 5. Domain Layer (`internal/domain/`)

Zero external dependencies. Pure Go types and business rules.

### 5.1 Task Entity

```go
package entity

import (
    "time"
    "yourmod/internal/domain/valueobject"
    "yourmod/internal/domain/domainerror"
)

type Task struct {
    title       string
    description string
    dueDate     *time.Time
    priority    valueobject.Priority
    labels      []string
    checklist   []string
}

func NewTask(title string, opts ...TaskOption) (*Task, error) {
    if title == "" {
        return nil, domainerror.ErrEmptyTaskTitle
    }
    t := &Task{
        title:    title,
        priority: valueobject.PriorityMedium,
    }
    for _, opt := range opts {
        opt(t)
    }
    return t, nil
}

type TaskOption func(*Task)

func WithDescription(desc string) TaskOption {
    return func(t *Task) { t.description = desc }
}

func WithDueDate(d time.Time) TaskOption {
    return func(t *Task) { t.dueDate = &d }
}

func WithPriority(p valueobject.Priority) TaskOption {
    return func(t *Task) { t.priority = p }
}

func WithLabels(labels []string) TaskOption {
    return func(t *Task) { t.labels = labels }
}

func WithChecklist(items []string) TaskOption {
    return func(t *Task) { t.checklist = items }
}

func (t *Task) Title() string                  { return t.title }
func (t *Task) Description() string            { return t.description }
func (t *Task) DueDate() *time.Time            { return t.dueDate }
func (t *Task) Priority() valueobject.Priority { return t.priority }
func (t *Task) Labels() []string               { return t.labels }
func (t *Task) Checklist() []string            { return t.checklist }
func (t *Task) IsHighPriority() bool           { return t.priority == valueobject.PriorityHigh }
```

### 5.2 Priority Value Object

```go
package valueobject

import "yourmod/internal/domain/domainerror"

type Priority string

const (
    PriorityLow    Priority = "low"
    PriorityMedium Priority = "medium"
    PriorityHigh   Priority = "high"
)

func NewPriority(s string) (Priority, error) {
    switch s {
    case "low":
        return PriorityLow, nil
    case "medium", "":
        return PriorityMedium, nil
    case "high":
        return PriorityHigh, nil
    default:
        return "", domainerror.ErrInvalidPriority
    }
}

func (p Priority) String() string { return string(p) }
```

### 5.3 Telegram ID Value Object

```go
package valueobject

type TelegramID int64

func NewTelegramID(id int64) TelegramID { return TelegramID(id) }
func (t TelegramID) Int64() int64       { return int64(t) }
```

### 5.4 Domain Errors

```go
package domainerror

import "errors"

var (
    ErrEmptyTaskTitle  = errors.New("task title cannot be empty")
    ErrInvalidPriority = errors.New("invalid priority value")
    ErrBoardNotSet     = errors.New("default board not configured")
    ErrListNotSet      = errors.New("default list not configured")
    ErrUserNotFound    = errors.New("user not found")
    ErrParsingFailed   = errors.New("failed to parse message into task")
    ErrCardCreation    = errors.New("failed to create card on board")
)
```

### 5.5 User Entity

```go
package entity

import "yourmod/internal/domain/valueobject"

type User struct {
    telegramID   valueobject.TelegramID
    trelloToken  string
    defaultBoard string
    defaultList  string
    useLLM       bool
}

func NewUser(telegramID valueobject.TelegramID) *User {
    return &User{telegramID: telegramID, useLLM: true}
}

func (u *User) TelegramID() valueobject.TelegramID { return u.telegramID }
func (u *User) TrelloToken() string                 { return u.trelloToken }
func (u *User) DefaultBoard() string                { return u.defaultBoard }
func (u *User) DefaultList() string                 { return u.defaultList }
func (u *User) HasBoardConfigured() bool            { return u.defaultBoard != "" }
func (u *User) HasListConfigured() bool             { return u.defaultList != "" }
func (u *User) UseLLM() bool                        { return u.useLLM }

func (u *User) SetTrelloToken(token string)    { u.trelloToken = token }
func (u *User) SetDefaultBoard(boardID string) { u.defaultBoard = boardID }
func (u *User) SetDefaultList(listID string)   { u.defaultList = listID }
func (u *User) SetUseLLM(use bool)             { u.useLLM = use }
```

### 5.6 Board Entity

```go
package entity

type Board struct {
    id   string
    name string
}

func NewBoard(id, name string) *Board { return &Board{id: id, name: name} }
func (b *Board) ID() string           { return b.id }
func (b *Board) Name() string         { return b.name }

type BoardList struct {
    id   string
    name string
}

func NewBoardList(id, name string) *BoardList { return &BoardList{id: id, name: name} }
func (l *BoardList) ID() string               { return l.id }
func (l *BoardList) Name() string             { return l.name }
```

---

## 6. Use Case Layer (`internal/usecase/`)

Depends ONLY on domain layer. Defines ports that infrastructure must implement.

### 6.1 Ports

```go
// port/task_parser.go
package port

import (
    "context"
    "yourmod/internal/domain/entity"
)

type TaskParser interface {
    Parse(ctx context.Context, rawMessage string) (*entity.Task, error)
}
```

```go
// port/task_board.go
package port

import "context"

type BoardInfo struct {
    ID   string
    Name string
}

type ListInfo struct {
    ID   string
    Name string
}

type LabelInfo struct {
    ID    string
    Name  string
    Color string
}

type CardResult struct {
    CardID  string
    CardURL string
}

type CreateCardParams struct {
    ListID      string
    Title       string
    Description string
    DueDate     *string
    LabelIDs    []string
    Position    string
}

type TaskBoard interface {
    GetBoards(ctx context.Context, token string) ([]BoardInfo, error)
    GetLists(ctx context.Context, token string, boardID string) ([]ListInfo, error)
    GetLabels(ctx context.Context, token string, boardID string) ([]LabelInfo, error)
    CreateCard(ctx context.Context, token string, params CreateCardParams) (*CardResult, error)
}
```

```go
// port/user_repository.go
package port

import (
    "context"
    "yourmod/internal/domain/entity"
    "yourmod/internal/domain/valueobject"
)

type UserRepository interface {
    FindByTelegramID(ctx context.Context, id valueobject.TelegramID) (*entity.User, error)
    Save(ctx context.Context, user *entity.User) error
}
```

```go
// port/task_log_repository.go
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
```

### 6.2 DTOs

```go
// usecase/dto/create_task_input.go
package dto

type CreateTaskInput struct {
    TelegramID int64
    RawMessage string
}
```

```go
// usecase/dto/create_task_output.go
package dto

import "time"

type CreateTaskOutput struct {
    CardURL   string
    TaskTitle string
    DueDate   *time.Time
    Priority  string
    Labels    []string
}
```

```go
// usecase/dto/list_boards_output.go
package dto

type ListBoardsOutput struct {
    Boards []BoardItem
}

type BoardItem struct {
    ID   string
    Name string
}
```

### 6.3 CreateTask Use Case

```go
package usecase

import (
    "context"
    "fmt"
    "yourmod/internal/domain/domainerror"
    "yourmod/internal/domain/valueobject"
    "yourmod/internal/usecase/dto"
    "yourmod/internal/usecase/port"
)

type CreateTaskUseCase struct {
    parser   port.TaskParser
    board    port.TaskBoard
    userRepo port.UserRepository
    taskLog  port.TaskLogRepository
}

func NewCreateTaskUseCase(
    parser port.TaskParser,
    board port.TaskBoard,
    userRepo port.UserRepository,
    taskLog port.TaskLogRepository,
) *CreateTaskUseCase {
    return &CreateTaskUseCase{
        parser: parser, board: board,
        userRepo: userRepo, taskLog: taskLog,
    }
}

func (uc *CreateTaskUseCase) Execute(
    ctx context.Context,
    input dto.CreateTaskInput,
) (*dto.CreateTaskOutput, error) {
    // 1. Load user preferences
    user, err := uc.userRepo.FindByTelegramID(ctx, valueobject.TelegramID(input.TelegramID))
    if err != nil {
        return nil, fmt.Errorf("find user: %w", err)
    }
    if !user.HasBoardConfigured() {
        return nil, domainerror.ErrBoardNotSet
    }
    if !user.HasListConfigured() {
        return nil, domainerror.ErrListNotSet
    }

    // 2. Parse message into domain Task entity
    task, err := uc.parser.Parse(ctx, input.RawMessage)
    if err != nil {
        return nil, fmt.Errorf("%w: %v", domainerror.ErrParsingFailed, err)
    }

    // 3. Map domain entity to board card params
    var dueStr *string
    if task.DueDate() != nil {
        s := task.DueDate().Format("2006-01-02T15:04:05Z")
        dueStr = &s
    }

    position := "bottom"
    if task.IsHighPriority() {
        position = "top"
    }

    params := port.CreateCardParams{
        ListID:      user.DefaultList(),
        Title:       task.Title(),
        Description: task.Description(),
        DueDate:     dueStr,
        Position:    position,
    }

    // 4. Create card on board
    result, err := uc.board.CreateCard(ctx, user.TrelloToken(), params)
    if err != nil {
        return nil, fmt.Errorf("%w: %v", domainerror.ErrCardCreation, err)
    }

    // 5. Log task creation
    _ = uc.taskLog.Log(ctx, port.TaskLogEntry{
        TelegramID: input.TelegramID,
        Message:    input.RawMessage,
        CardID:     result.CardID,
    })

    // 6. Return output DTO
    return &dto.CreateTaskOutput{
        CardURL:   result.CardURL,
        TaskTitle: task.Title(),
        DueDate:   task.DueDate(),
        Priority:  string(task.Priority()),
        Labels:    task.Labels(),
    }, nil
}
```

### 6.4 ListBoards Use Case

```go
package usecase

import (
    "context"
    "fmt"
    "yourmod/internal/domain/valueobject"
    "yourmod/internal/usecase/dto"
    "yourmod/internal/usecase/port"
)

type ListBoardsUseCase struct {
    board    port.TaskBoard
    userRepo port.UserRepository
}

func NewListBoardsUseCase(board port.TaskBoard, userRepo port.UserRepository) *ListBoardsUseCase {
    return &ListBoardsUseCase{board: board, userRepo: userRepo}
}

func (uc *ListBoardsUseCase) Execute(ctx context.Context, telegramID int64) (*dto.ListBoardsOutput, error) {
    user, err := uc.userRepo.FindByTelegramID(ctx, valueobject.TelegramID(telegramID))
    if err != nil {
        return nil, fmt.Errorf("find user: %w", err)
    }

    boards, err := uc.board.GetBoards(ctx, user.TrelloToken())
    if err != nil {
        return nil, fmt.Errorf("get boards: %w", err)
    }

    items := make([]dto.BoardItem, len(boards))
    for i, b := range boards {
        items[i] = dto.BoardItem{ID: b.ID, Name: b.Name}
    }
    return &dto.ListBoardsOutput{Boards: items}, nil
}
```

### 6.5 SelectBoard & SelectList Use Cases

```go
// usecase/select_board.go
package usecase

import (
    "context"
    "yourmod/internal/domain/valueobject"
    "yourmod/internal/usecase/port"
)

type SelectBoardUseCase struct {
    userRepo port.UserRepository
}

func NewSelectBoardUseCase(userRepo port.UserRepository) *SelectBoardUseCase {
    return &SelectBoardUseCase{userRepo: userRepo}
}

func (uc *SelectBoardUseCase) Execute(ctx context.Context, telegramID int64, boardID string) error {
    user, err := uc.userRepo.FindByTelegramID(ctx, valueobject.TelegramID(telegramID))
    if err != nil {
        return err
    }
    user.SetDefaultBoard(boardID)
    return uc.userRepo.Save(ctx, user)
}
```

```go
// usecase/select_list.go
package usecase

import (
    "context"
    "yourmod/internal/domain/valueobject"
    "yourmod/internal/usecase/port"
)

type SelectListUseCase struct {
    userRepo port.UserRepository
}

func NewSelectListUseCase(userRepo port.UserRepository) *SelectListUseCase {
    return &SelectListUseCase{userRepo: userRepo}
}

func (uc *SelectListUseCase) Execute(ctx context.Context, telegramID int64, listID string) error {
    user, err := uc.userRepo.FindByTelegramID(ctx, valueobject.TelegramID(telegramID))
    if err != nil {
        return err
    }
    user.SetDefaultList(listID)
    return uc.userRepo.Save(ctx, user)
}
```

### 6.6 Unit Test (CreateTask with Mocks)

```go
package usecase_test

import (
    "context"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"

    "yourmod/internal/domain/entity"
    "yourmod/internal/domain/domainerror"
    "yourmod/internal/domain/valueobject"
    "yourmod/internal/usecase"
    "yourmod/internal/usecase/dto"
    "yourmod/internal/usecase/port"
)

type MockParser struct{ mock.Mock }
func (m *MockParser) Parse(ctx context.Context, msg string) (*entity.Task, error) {
    args := m.Called(ctx, msg)
    if args.Get(0) == nil { return nil, args.Error(1) }
    return args.Get(0).(*entity.Task), args.Error(1)
}

type MockBoard struct{ mock.Mock }
func (m *MockBoard) GetBoards(ctx context.Context, token string) ([]port.BoardInfo, error) {
    args := m.Called(ctx, token)
    return args.Get(0).([]port.BoardInfo), args.Error(1)
}
func (m *MockBoard) GetLists(ctx context.Context, token string, boardID string) ([]port.ListInfo, error) {
    args := m.Called(ctx, token, boardID)
    return args.Get(0).([]port.ListInfo), args.Error(1)
}
func (m *MockBoard) GetLabels(ctx context.Context, token string, boardID string) ([]port.LabelInfo, error) {
    args := m.Called(ctx, token, boardID)
    return args.Get(0).([]port.LabelInfo), args.Error(1)
}
func (m *MockBoard) CreateCard(ctx context.Context, token string, p port.CreateCardParams) (*port.CardResult, error) {
    args := m.Called(ctx, token, p)
    if args.Get(0) == nil { return nil, args.Error(1) }
    return args.Get(0).(*port.CardResult), args.Error(1)
}

type MockUserRepo struct{ mock.Mock }
func (m *MockUserRepo) FindByTelegramID(ctx context.Context, id valueobject.TelegramID) (*entity.User, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil { return nil, args.Error(1) }
    return args.Get(0).(*entity.User), args.Error(1)
}
func (m *MockUserRepo) Save(ctx context.Context, user *entity.User) error {
    return m.Called(ctx, user).Error(0)
}

type MockTaskLog struct{ mock.Mock }
func (m *MockTaskLog) Log(ctx context.Context, entry port.TaskLogEntry) error {
    return m.Called(ctx, entry).Error(0)
}

func TestCreateTask_Success(t *testing.T) {
    parser := new(MockParser)
    board := new(MockBoard)
    userRepo := new(MockUserRepo)
    taskLog := new(MockTaskLog)

    uc := usecase.NewCreateTaskUseCase(parser, board, userRepo, taskLog)

    user := entity.NewUser(valueobject.TelegramID(12345))
    user.SetDefaultBoard("board-1")
    user.SetDefaultList("list-1")
    user.SetTrelloToken("trello-token-xyz")
    userRepo.On("FindByTelegramID", mock.Anything, valueobject.TelegramID(12345)).Return(user, nil)

    due := time.Date(2025, 3, 7, 0, 0, 0, 0, time.UTC)
    task, _ := entity.NewTask("Fix payment bug",
        entity.WithPriority(valueobject.PriorityHigh),
        entity.WithDueDate(due),
        entity.WithLabels([]string{"backend"}),
    )
    parser.On("Parse", mock.Anything, mock.Anything).Return(task, nil)

    board.On("CreateCard", mock.Anything, "trello-token-xyz", mock.MatchedBy(func(p port.CreateCardParams) bool {
        return p.Title == "Fix payment bug" && p.Position == "top" && p.ListID == "list-1"
    })).Return(&port.CardResult{
        CardID: "card-123", CardURL: "https://trello.com/c/abc123",
    }, nil)

    taskLog.On("Log", mock.Anything, mock.Anything).Return(nil)

    output, err := uc.Execute(context.Background(), dto.CreateTaskInput{
        TelegramID: 12345,
        RawMessage: "Fix payment bug, urgent, due Friday #backend",
    })

    assert.NoError(t, err)
    assert.Equal(t, "Fix payment bug", output.TaskTitle)
    assert.Equal(t, "https://trello.com/c/abc123", output.CardURL)
    assert.Equal(t, "high", output.Priority)
    assert.Equal(t, []string{"backend"}, output.Labels)
    board.AssertExpectations(t)
}

func TestCreateTask_BoardNotConfigured(t *testing.T) {
    parser := new(MockParser)
    board := new(MockBoard)
    userRepo := new(MockUserRepo)
    taskLog := new(MockTaskLog)

    uc := usecase.NewCreateTaskUseCase(parser, board, userRepo, taskLog)

    user := entity.NewUser(valueobject.TelegramID(12345))
    userRepo.On("FindByTelegramID", mock.Anything, mock.Anything).Return(user, nil)

    _, err := uc.Execute(context.Background(), dto.CreateTaskInput{
        TelegramID: 12345, RawMessage: "Some task",
    })

    assert.ErrorIs(t, err, domainerror.ErrBoardNotSet)
    parser.AssertNotCalled(t, "Parse")
}
```

---

## 7. Adapter Layer (`internal/adapter/`)

### 7.1 Telegram Controller

```go
package controller

import (
    "context"
    "yourmod/internal/usecase"
    "yourmod/internal/usecase/dto"
)

type TelegramController struct {
    createTask  *usecase.CreateTaskUseCase
    listBoards  *usecase.ListBoardsUseCase
    selectBoard *usecase.SelectBoardUseCase
    selectList  *usecase.SelectListUseCase
}

func NewTelegramController(
    createTask *usecase.CreateTaskUseCase,
    listBoards *usecase.ListBoardsUseCase,
    selectBoard *usecase.SelectBoardUseCase,
    selectList *usecase.SelectListUseCase,
) *TelegramController {
    return &TelegramController{
        createTask: createTask, listBoards: listBoards,
        selectBoard: selectBoard, selectList: selectList,
    }
}

func (c *TelegramController) HandleMessage(ctx context.Context, telegramID int64, text string) (*dto.CreateTaskOutput, error) {
    return c.createTask.Execute(ctx, dto.CreateTaskInput{TelegramID: telegramID, RawMessage: text})
}

func (c *TelegramController) HandleListBoards(ctx context.Context, telegramID int64) (*dto.ListBoardsOutput, error) {
    return c.listBoards.Execute(ctx, telegramID)
}

func (c *TelegramController) HandleSelectBoard(ctx context.Context, telegramID int64, boardID string) error {
    return c.selectBoard.Execute(ctx, telegramID, boardID)
}

func (c *TelegramController) HandleSelectList(ctx context.Context, telegramID int64, listID string) error {
    return c.selectList.Execute(ctx, telegramID, listID)
}
```

### 7.2 Telegram Presenter

```go
package presenter

import (
    "fmt"
    "strings"
    "yourmod/internal/usecase/dto"
)

type TelegramPresenter struct{}

func NewTelegramPresenter() *TelegramPresenter {
    return &TelegramPresenter{}
}

func (p *TelegramPresenter) FormatTaskCreated(output *dto.CreateTaskOutput) string {
    var sb strings.Builder
    sb.WriteString("✅ *Task Created!*\n\n")
    sb.WriteString(fmt.Sprintf("*Title:* %s\n", output.TaskTitle))
    sb.WriteString(fmt.Sprintf("*Priority:* %s\n", output.Priority))
    if output.DueDate != nil {
        sb.WriteString(fmt.Sprintf("*Due:* %s\n", output.DueDate.Format("Jan 2, 2006")))
    }
    if len(output.Labels) > 0 {
        sb.WriteString(fmt.Sprintf("*Labels:* %s\n", strings.Join(output.Labels, ", ")))
    }
    sb.WriteString(fmt.Sprintf("\n🔗 %s", output.CardURL))
    return sb.String()
}

func (p *TelegramPresenter) FormatTaskPreview(output *dto.CreateTaskOutput) string {
    var sb strings.Builder
    sb.WriteString("📋 *Create this task?*\n\n")
    sb.WriteString(fmt.Sprintf("*Title:* %s\n", output.TaskTitle))
    sb.WriteString(fmt.Sprintf("*Priority:* %s\n", output.Priority))
    if output.DueDate != nil {
        sb.WriteString(fmt.Sprintf("*Due:* %s\n", output.DueDate.Format("Jan 2, 2006")))
    }
    if len(output.Labels) > 0 {
        sb.WriteString(fmt.Sprintf("*Labels:* %s\n", strings.Join(output.Labels, ", ")))
    }
    return sb.String()
}

func (p *TelegramPresenter) FormatBoardList(output *dto.ListBoardsOutput) string {
    var sb strings.Builder
    sb.WriteString("📌 *Your Trello Boards:*\n\n")
    for i, b := range output.Boards {
        sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, b.Name))
    }
    sb.WriteString("\nTap a board to set it as default.")
    return sb.String()
}

func (p *TelegramPresenter) FormatError(err error) string {
    return fmt.Sprintf("❌ Something went wrong: %s", err.Error())
}

func (p *TelegramPresenter) FormatHelp() string {
    return `🤖 *Telegram → Trello Bot*

Just send me any message and I'll create a Trello card!

*Tips:*
• Include "urgent" or "high priority" to mark as important
• Add "#label" to tag your task
• Say "due Friday" or "by March 15" to set a deadline

*Commands:*
/boards — List your Trello boards
/setboard — Set default board
/setlist — Set default list
/help — Show this message`
}
```

### 7.3 Trello Gateway

```go
package gateway

import (
    "context"
    "yourmod/internal/infrastructure/trello"
    "yourmod/internal/usecase/port"
)

type TrelloGateway struct {
    client *trello.Client
}

func NewTrelloGateway(client *trello.Client) *TrelloGateway {
    return &TrelloGateway{client: client}
}

func (g *TrelloGateway) GetBoards(ctx context.Context, token string) ([]port.BoardInfo, error) {
    boards, err := g.client.GetBoards(ctx, token)
    if err != nil {
        return nil, err
    }
    result := make([]port.BoardInfo, len(boards))
    for i, b := range boards {
        result[i] = port.BoardInfo{ID: b.ID, Name: b.Name}
    }
    return result, nil
}

func (g *TrelloGateway) GetLists(ctx context.Context, token string, boardID string) ([]port.ListInfo, error) {
    lists, err := g.client.GetLists(ctx, token, boardID)
    if err != nil {
        return nil, err
    }
    result := make([]port.ListInfo, len(lists))
    for i, l := range lists {
        result[i] = port.ListInfo{ID: l.ID, Name: l.Name}
    }
    return result, nil
}

func (g *TrelloGateway) GetLabels(ctx context.Context, token string, boardID string) ([]port.LabelInfo, error) {
    labels, err := g.client.GetLabels(ctx, token, boardID)
    if err != nil {
        return nil, err
    }
    result := make([]port.LabelInfo, len(labels))
    for i, l := range labels {
        result[i] = port.LabelInfo{ID: l.ID, Name: l.Name, Color: l.Color}
    }
    return result, nil
}

func (g *TrelloGateway) CreateCard(ctx context.Context, token string, params port.CreateCardParams) (*port.CardResult, error) {
    trelloParams := trello.CreateCardRequest{
        Name:        params.Title,
        Description: params.Description,
        ListID:      params.ListID,
        Position:    params.Position,
    }
    if params.DueDate != nil {
        trelloParams.Due = *params.DueDate
    }
    if len(params.LabelIDs) > 0 {
        trelloParams.LabelIDs = params.LabelIDs
    }

    resp, err := g.client.CreateCard(ctx, token, trelloParams)
    if err != nil {
        return nil, err
    }
    return &port.CardResult{CardID: resp.ID, CardURL: resp.ShortURL}, nil
}
```

### 7.4 Claude Parser Gateway

```go
package gateway

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "yourmod/internal/domain/entity"
    "yourmod/internal/domain/valueobject"
    "yourmod/internal/infrastructure/claude"
)

type ClaudeParserGateway struct {
    client *claude.Client
}

func NewClaudeParserGateway(client *claude.Client) *ClaudeParserGateway {
    return &ClaudeParserGateway{client: client}
}

const parsePrompt = `You are a task parser. Extract structured task data from the user's message.
Return ONLY valid JSON with these fields:
{
  "title": "string (required, concise task title)",
  "description": "string (optional, additional details)",
  "due_date": "string (optional, ISO 8601 date like 2025-03-07)",
  "priority": "string (low|medium|high, default medium)",
  "labels": ["string array (optional, extracted tags/categories)"],
  "checklist": ["string array (optional, subtasks if mentioned)"]
}

Today's date is %s. Resolve relative dates like "tomorrow", "next Friday" relative to today.`

type parsedTaskJSON struct {
    Title       string   `json:"title"`
    Description string   `json:"description"`
    DueDate     string   `json:"due_date"`
    Priority    string   `json:"priority"`
    Labels      []string `json:"labels"`
    Checklist   []string `json:"checklist"`
}

func (g *ClaudeParserGateway) Parse(ctx context.Context, rawMessage string) (*entity.Task, error) {
    prompt := fmt.Sprintf(parsePrompt, time.Now().Format("2006-01-02"))

    response, err := g.client.SendMessage(ctx, prompt, rawMessage)
    if err != nil {
        return nil, fmt.Errorf("claude API: %w", err)
    }

    var parsed parsedTaskJSON
    if err := json.Unmarshal([]byte(response), &parsed); err != nil {
        return nil, fmt.Errorf("parse claude response: %w", err)
    }

    opts := []entity.TaskOption{}
    if parsed.Description != "" {
        opts = append(opts, entity.WithDescription(parsed.Description))
    }
    if parsed.DueDate != "" {
        if due, err := time.Parse("2006-01-02", parsed.DueDate); err == nil {
            opts = append(opts, entity.WithDueDate(due))
        }
    }
    if parsed.Priority != "" {
        if p, err := valueobject.NewPriority(parsed.Priority); err == nil {
            opts = append(opts, entity.WithPriority(p))
        }
    }
    if len(parsed.Labels) > 0 {
        opts = append(opts, entity.WithLabels(parsed.Labels))
    }
    if len(parsed.Checklist) > 0 {
        opts = append(opts, entity.WithChecklist(parsed.Checklist))
    }

    return entity.NewTask(parsed.Title, opts...)
}
```

### 7.5 Rule-Based Parser Gateway

```go
package gateway

import (
    "context"
    "fmt"
    "regexp"
    "strings"
    "time"

    "yourmod/internal/domain/entity"
    "yourmod/internal/domain/valueobject"
)

type RuleParserGateway struct{}

func NewRuleParserGateway() *RuleParserGateway {
    return &RuleParserGateway{}
}

var (
    dueDateRegex  = regexp.MustCompile(`(?i)\b(?:due|by)\s+(\w+(?:\s+\d{1,2})?(?:,?\s*\d{4})?)`)
    priorityRegex = regexp.MustCompile(`(?i)\b(urgent|high\s*priority|low\s*priority)\b`)
    labelRegex    = regexp.MustCompile(`#(\w+)`)
)

func (g *RuleParserGateway) Parse(ctx context.Context, rawMessage string) (*entity.Task, error) {
    opts := []entity.TaskOption{}

    // Extract priority
    if match := priorityRegex.FindString(rawMessage); match != "" {
        match = strings.ToLower(match)
        if strings.Contains(match, "urgent") || strings.Contains(match, "high") {
            opts = append(opts, entity.WithPriority(valueobject.PriorityHigh))
        } else if strings.Contains(match, "low") {
            opts = append(opts, entity.WithPriority(valueobject.PriorityLow))
        }
    }

    // Extract labels
    labelMatches := labelRegex.FindAllStringSubmatch(rawMessage, -1)
    if len(labelMatches) > 0 {
        labels := make([]string, len(labelMatches))
        for i, m := range labelMatches {
            labels[i] = m[1]
        }
        opts = append(opts, entity.WithLabels(labels))
    }

    // Extract due date
    if match := dueDateRegex.FindStringSubmatch(rawMessage); len(match) > 1 {
        if due, err := parseNaturalDate(match[1]); err == nil {
            opts = append(opts, entity.WithDueDate(due))
        }
    }

    // Title: remove extracted patterns
    title := rawMessage
    title = dueDateRegex.ReplaceAllString(title, "")
    title = priorityRegex.ReplaceAllString(title, "")
    title = labelRegex.ReplaceAllString(title, "")
    title = strings.TrimRight(strings.TrimSpace(strings.Join(strings.Fields(title), " ")), " ,.")

    return entity.NewTask(title, opts...)
}

func parseNaturalDate(s string) (time.Time, error) {
    s = strings.ToLower(strings.TrimSpace(s))
    now := time.Now()

    switch s {
    case "today":
        return now.Truncate(24 * time.Hour), nil
    case "tomorrow":
        return now.AddDate(0, 0, 1).Truncate(24 * time.Hour), nil
    }

    weekdays := map[string]time.Weekday{
        "monday": time.Monday, "tuesday": time.Tuesday,
        "wednesday": time.Wednesday, "thursday": time.Thursday,
        "friday": time.Friday, "saturday": time.Saturday, "sunday": time.Sunday,
    }
    if wd, ok := weekdays[s]; ok {
        daysUntil := int(wd - now.Weekday())
        if daysUntil <= 0 {
            daysUntil += 7
        }
        return now.AddDate(0, 0, daysUntil).Truncate(24 * time.Hour), nil
    }

    formats := []string{"2006-01-02", "January 2", "Jan 2", "January 2, 2006", "Jan 2, 2006"}
    for _, f := range formats {
        if t, err := time.Parse(f, s); err == nil {
            if t.Year() == 0 {
                t = t.AddDate(now.Year(), 0, 0)
            }
            return t, nil
        }
    }
    return time.Time{}, fmt.Errorf("cannot parse date: %s", s)
}
```

### 7.6 Parser Chain Gateway

```go
package gateway

import (
    "context"
    "log/slog"
    "yourmod/internal/domain/entity"
    "yourmod/internal/usecase/port"
)

type ParserChainGateway struct {
    primary  port.TaskParser
    fallback port.TaskParser
    logger   *slog.Logger
}

func NewParserChainGateway(primary, fallback port.TaskParser, logger *slog.Logger) *ParserChainGateway {
    return &ParserChainGateway{primary: primary, fallback: fallback, logger: logger}
}

func (g *ParserChainGateway) Parse(ctx context.Context, msg string) (*entity.Task, error) {
    task, err := g.primary.Parse(ctx, msg)
    if err != nil {
        g.logger.Warn("primary parser failed, falling back", "error", err)
        return g.fallback.Parse(ctx, msg)
    }
    return task, nil
}
```

---

## 8. Infrastructure Layer (`internal/infrastructure/`)

### 8.1 Trello HTTP Client

```go
package trello

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "net/url"
    "strings"
    "time"
)

const baseURL = "https://api.trello.com"

type Client struct {
    apiKey     string
    httpClient *http.Client
}

func NewClient(apiKey string) *Client {
    return &Client{
        apiKey:     apiKey,
        httpClient: &http.Client{Timeout: 10 * time.Second},
    }
}

type CreateCardRequest struct {
    Name        string
    Description string
    ListID      string
    Due         string
    LabelIDs    []string
    Position    string
}

type CardResponse struct {
    ID       string `json:"id"`
    Name     string `json:"name"`
    ShortURL string `json:"shortUrl"`
    URL      string `json:"url"`
}

type BoardResponse struct {
    ID   string `json:"id"`
    Name string `json:"name"`
}

type ListResponse struct {
    ID   string `json:"id"`
    Name string `json:"name"`
}

type LabelResponse struct {
    ID    string `json:"id"`
    Name  string `json:"name"`
    Color string `json:"color"`
}

func (c *Client) GetBoards(ctx context.Context, token string) ([]BoardResponse, error) {
    endpoint := fmt.Sprintf("%s/1/members/me/boards?key=%s&token=%s&fields=id,name",
        baseURL, c.apiKey, token)
    var boards []BoardResponse
    if err := c.doGet(ctx, endpoint, &boards); err != nil {
        return nil, fmt.Errorf("get boards: %w", err)
    }
    return boards, nil
}

func (c *Client) GetLists(ctx context.Context, token string, boardID string) ([]ListResponse, error) {
    endpoint := fmt.Sprintf("%s/1/boards/%s/lists?key=%s&token=%s&fields=id,name",
        baseURL, boardID, c.apiKey, token)
    var lists []ListResponse
    if err := c.doGet(ctx, endpoint, &lists); err != nil {
        return nil, fmt.Errorf("get lists: %w", err)
    }
    return lists, nil
}

func (c *Client) GetLabels(ctx context.Context, token string, boardID string) ([]LabelResponse, error) {
    endpoint := fmt.Sprintf("%s/1/boards/%s/labels?key=%s&token=%s",
        baseURL, boardID, c.apiKey, token)
    var labels []LabelResponse
    if err := c.doGet(ctx, endpoint, &labels); err != nil {
        return nil, fmt.Errorf("get labels: %w", err)
    }
    return labels, nil
}

func (c *Client) CreateCard(ctx context.Context, token string, req CreateCardRequest) (*CardResponse, error) {
    body := url.Values{}
    body.Set("key", c.apiKey)
    body.Set("token", token)
    body.Set("name", req.Name)
    body.Set("desc", req.Description)
    body.Set("idList", req.ListID)
    if req.Due != "" {
        body.Set("due", req.Due)
    }
    if len(req.LabelIDs) > 0 {
        body.Set("idLabels", strings.Join(req.LabelIDs, ","))
    }
    if req.Position != "" {
        body.Set("pos", req.Position)
    }

    httpReq, err := http.NewRequestWithContext(ctx, "POST",
        baseURL+"/1/cards", strings.NewReader(body.Encode()))
    if err != nil {
        return nil, fmt.Errorf("build request: %w", err)
    }
    httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

    resp, err := c.httpClient.Do(httpReq)
    if err != nil {
        return nil, fmt.Errorf("execute request: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("trello API returned status %d", resp.StatusCode)
    }

    var card CardResponse
    if err := json.NewDecoder(resp.Body).Decode(&card); err != nil {
        return nil, fmt.Errorf("decode response: %w", err)
    }
    return &card, nil
}

func (c *Client) doGet(ctx context.Context, url string, target interface{}) error {
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return err
    }
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("HTTP %d", resp.StatusCode)
    }
    return json.NewDecoder(resp.Body).Decode(target)
}
```

### 8.2 Claude HTTP Client

```go
package claude

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
)

const claudeBaseURL = "https://api.anthropic.com/v1/messages"

type Client struct {
    apiKey     string
    model      string
    httpClient *http.Client
}

func NewClient(apiKey, model string) *Client {
    return &Client{
        apiKey:     apiKey,
        model:      model,
        httpClient: &http.Client{Timeout: 30 * time.Second},
    }
}

type messageRequest struct {
    Model     string    `json:"model"`
    MaxTokens int       `json:"max_tokens"`
    System    string    `json:"system"`
    Messages  []message `json:"messages"`
}

type message struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

type messageResponse struct {
    Content []struct {
        Text string `json:"text"`
    } `json:"content"`
}

func (c *Client) SendMessage(ctx context.Context, systemPrompt, userMessage string) (string, error) {
    reqBody := messageRequest{
        Model:     c.model,
        MaxTokens: 1024,
        System:    systemPrompt,
        Messages:  []message{{Role: "user", Content: userMessage}},
    }

    bodyBytes, err := json.Marshal(reqBody)
    if err != nil {
        return "", fmt.Errorf("marshal request: %w", err)
    }

    req, err := http.NewRequestWithContext(ctx, "POST", claudeBaseURL, bytes.NewReader(bodyBytes))
    if err != nil {
        return "", fmt.Errorf("build request: %w", err)
    }
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("x-api-key", c.apiKey)
    req.Header.Set("anthropic-version", "2023-06-01")

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return "", fmt.Errorf("execute request: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return "", fmt.Errorf("claude API returned status %d", resp.StatusCode)
    }

    var msgResp messageResponse
    if err := json.NewDecoder(resp.Body).Decode(&msgResp); err != nil {
        return "", fmt.Errorf("decode response: %w", err)
    }
    if len(msgResp.Content) == 0 {
        return "", fmt.Errorf("empty response from claude")
    }
    return msgResp.Content[0].Text, nil
}
```

### 8.3 SQLite Persistence

```go
// infrastructure/persistence/sqlite.go
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
```

```go
// infrastructure/persistence/user_repo_sqlite.go
package persistence

import (
    "context"
    "database/sql"
    "yourmod/internal/domain/entity"
    "yourmod/internal/domain/domainerror"
    "yourmod/internal/domain/valueobject"
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
```

```go
// infrastructure/persistence/task_log_repo_sqlite.go
package persistence

import (
    "context"
    "database/sql"
    "yourmod/internal/usecase/port"
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
```

### 8.4 Telegram Bot & Router

```go
// infrastructure/telegram/bot.go
package telegram

import (
    "log/slog"
    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
    api    *tgbotapi.BotAPI
    router *Router
    logger *slog.Logger
}

func NewBot(token string, router *Router, logger *slog.Logger) (*Bot, error) {
    api, err := tgbotapi.NewBotAPI(token)
    if err != nil {
        return nil, err
    }
    return &Bot{api: api, router: router, logger: logger}, nil
}

func (b *Bot) StartPolling() {
    b.logger.Info("bot started", "username", b.api.Self.UserName)
    u := tgbotapi.NewUpdate(0)
    u.Timeout = 60
    updates := b.api.GetUpdatesChan(u)
    for update := range updates {
        go b.router.Route(b.api, update)
    }
}
```

```go
// infrastructure/telegram/router.go
package telegram

import (
    "context"
    "log/slog"

    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
    "yourmod/internal/adapter/controller"
    "yourmod/internal/adapter/presenter"
)

type Router struct {
    ctrl      *controller.TelegramController
    presenter *presenter.TelegramPresenter
    logger    *slog.Logger
}

func NewRouter(ctrl *controller.TelegramController, pres *presenter.TelegramPresenter, logger *slog.Logger) *Router {
    return &Router{ctrl: ctrl, presenter: pres, logger: logger}
}

func (r *Router) Route(api *tgbotapi.BotAPI, update tgbotapi.Update) {
    ctx := context.Background()

    if update.Message == nil {
        return
    }

    chatID := update.Message.Chat.ID
    userID := update.Message.From.ID

    if update.Message.IsCommand() {
        switch update.Message.Command() {
        case "start":
            r.sendText(api, chatID, "👋 Welcome! Send /help to get started.")
        case "help":
            r.sendText(api, chatID, r.presenter.FormatHelp())
        case "boards":
            output, err := r.ctrl.HandleListBoards(ctx, userID)
            if err != nil {
                r.sendText(api, chatID, r.presenter.FormatError(err))
                return
            }
            r.sendText(api, chatID, r.presenter.FormatBoardList(output))
        default:
            r.sendText(api, chatID, "Unknown command. Try /help")
        }
        return
    }

    output, err := r.ctrl.HandleMessage(ctx, userID, update.Message.Text)
    if err != nil {
        r.logger.Error("create task failed", "error", err, "user_id", userID)
        r.sendText(api, chatID, r.presenter.FormatError(err))
        return
    }
    r.sendText(api, chatID, r.presenter.FormatTaskCreated(output))
}

func (r *Router) sendText(api *tgbotapi.BotAPI, chatID int64, text string) {
    msg := tgbotapi.NewMessage(chatID, text)
    msg.ParseMode = "Markdown"
    if _, err := api.Send(msg); err != nil {
        r.logger.Error("failed to send message", "error", err, "chat_id", chatID)
    }
}
```

### 8.5 Inline Keyboards

```go
package telegram

import (
    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
    "yourmod/internal/usecase/port"
)

func BuildBoardKeyboard(boards []port.BoardInfo) tgbotapi.InlineKeyboardMarkup {
    var rows [][]tgbotapi.InlineKeyboardButton
    for _, b := range boards {
        rows = append(rows, tgbotapi.NewInlineKeyboardRow(
            tgbotapi.NewInlineKeyboardButtonData(b.Name, "board:"+b.ID),
        ))
    }
    return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func BuildListKeyboard(lists []port.ListInfo) tgbotapi.InlineKeyboardMarkup {
    var rows [][]tgbotapi.InlineKeyboardButton
    for _, l := range lists {
        rows = append(rows, tgbotapi.NewInlineKeyboardRow(
            tgbotapi.NewInlineKeyboardButtonData(l.Name, "list:"+l.ID),
        ))
    }
    return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func BuildConfirmKeyboard() tgbotapi.InlineKeyboardMarkup {
    return tgbotapi.NewInlineKeyboardMarkup(
        tgbotapi.NewInlineKeyboardRow(
            tgbotapi.NewInlineKeyboardButtonData("✅ Create", "confirm:create"),
            tgbotapi.NewInlineKeyboardButtonData("✏️ Edit", "confirm:edit"),
            tgbotapi.NewInlineKeyboardButtonData("❌ Cancel", "confirm:cancel"),
        ),
    )
}
```

### 8.6 Config

```go
package config

import (
    "log"
    "github.com/spf13/viper"
)

type Config struct {
    TelegramToken      string
    TelegramWebhookURL string
    TelegramMode       string
    TrelloAPIKey       string
    ClaudeAPIKey       string
    ClaudeModel        string
    DatabasePath       string
    Port               string
    LogLevel           string
}

func Load() *Config {
    viper.SetConfigFile(".env")
    viper.AutomaticEnv()
    if err := viper.ReadInConfig(); err != nil {
        log.Printf("no .env file found, using environment variables: %v", err)
    }
    return &Config{
        TelegramToken:      viper.GetString("TELEGRAM_BOT_TOKEN"),
        TelegramWebhookURL: viper.GetString("TELEGRAM_WEBHOOK_URL"),
        TelegramMode:       viper.GetString("TELEGRAM_MODE"),
        TrelloAPIKey:       viper.GetString("TRELLO_API_KEY"),
        ClaudeAPIKey:       viper.GetString("CLAUDE_API_KEY"),
        ClaudeModel:        viper.GetString("CLAUDE_MODEL"),
        DatabasePath:       viper.GetString("DATABASE_PATH"),
        Port:               viper.GetString("PORT"),
        LogLevel:           viper.GetString("LOG_LEVEL"),
    }
}
```

---

## 9. Composition Root (`cmd/bot/main.go`)

```go
package main

import (
    "context"
    "log/slog"
    "os"
    "os/signal"
    "syscall"

    "yourmod/internal/adapter/controller"
    "yourmod/internal/adapter/gateway"
    "yourmod/internal/adapter/presenter"
    "yourmod/internal/infrastructure/claude"
    "yourmod/internal/infrastructure/config"
    "yourmod/internal/infrastructure/persistence"
    "yourmod/internal/infrastructure/telegram"
    infraTrello "yourmod/internal/infrastructure/trello"
    "yourmod/internal/usecase"
)

func main() {
    cfg := config.Load()
    logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
        Level: parseLogLevel(cfg.LogLevel),
    }))

    // Infrastructure
    db, err := persistence.NewSQLiteDB(cfg.DatabasePath)
    if err != nil {
        logger.Error("failed to open database", "error", err)
        os.Exit(1)
    }
    defer db.Close()

    trelloClient := infraTrello.NewClient(cfg.TrelloAPIKey)
    claudeClient := claude.NewClient(cfg.ClaudeAPIKey, cfg.ClaudeModel)

    // Repositories
    userRepo := persistence.NewUserRepoSQLite(db)
    taskLogRepo := persistence.NewTaskLogRepoSQLite(db)

    // Gateways
    trelloGw := gateway.NewTrelloGateway(trelloClient)
    llmParser := gateway.NewClaudeParserGateway(claudeClient)
    ruleParser := gateway.NewRuleParserGateway()
    parserChain := gateway.NewParserChainGateway(llmParser, ruleParser, logger)

    // Use Cases
    createTaskUC := usecase.NewCreateTaskUseCase(parserChain, trelloGw, userRepo, taskLogRepo)
    listBoardsUC := usecase.NewListBoardsUseCase(trelloGw, userRepo)
    selectBoardUC := usecase.NewSelectBoardUseCase(userRepo)
    selectListUC := usecase.NewSelectListUseCase(userRepo)

    // Adapters
    ctrl := controller.NewTelegramController(createTaskUC, listBoardsUC, selectBoardUC, selectListUC)
    pres := presenter.NewTelegramPresenter()

    // Delivery
    router := telegram.NewRouter(ctrl, pres, logger)
    bot, err := telegram.NewBot(cfg.TelegramToken, router, logger)
    if err != nil {
        logger.Error("failed to create bot", "error", err)
        os.Exit(1)
    }

    // Graceful Shutdown
    ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    defer stop()

    go bot.StartPolling()
    <-ctx.Done()
    logger.Info("shutting down gracefully...")
}

func parseLogLevel(level string) slog.Level {
    switch level {
    case "debug":
        return slog.LevelDebug
    case "warn":
        return slog.LevelWarn
    case "error":
        return slog.LevelError
    default:
        return slog.LevelInfo
    }
}
```

---

## 10. Dependency Graph (Enforced)

```
main.go
  ├── imports infrastructure/*    ✅
  ├── imports adapter/*           ✅
  ├── imports usecase/*           ✅
  └── imports domain/*            ✅

infrastructure/*
  ├── imports adapter/*           ✅ (router calls controller)
  ├── imports usecase/port/*      ✅ (for port types)
  ├── imports domain/*            ✅
  └── imports usecase/*           ❌ NEVER

adapter/*
  ├── imports infrastructure/*    ✅ (gateway wraps infra client)
  ├── imports usecase/*           ✅ (controller calls use cases)
  └── imports domain/*            ✅

usecase/*
  ├── imports infrastructure/*    ❌ NEVER
  ├── imports adapter/*           ❌ NEVER
  └── imports domain/*            ✅ (only inward)

domain/*
  └── imports anything external   ❌ NEVER (zero dependencies)
```

---

## 11. `.env.example`

```env
# Telegram
TELEGRAM_BOT_TOKEN=your_bot_token_from_botfather
TELEGRAM_WEBHOOK_URL=https://yourdomain.com/webhook
TELEGRAM_MODE=polling

# Trello
TRELLO_API_KEY=your_trello_api_key

# Claude (optional)
CLAUDE_API_KEY=your_claude_api_key
CLAUDE_MODEL=claude-sonnet-4-5-20250929

# App
LOG_LEVEL=info
DATABASE_PATH=./data/bot.db
PORT=8080
```

---

## 12. Development Phases

### Phase 1 — Foundation + MVP (Week 1-2)

- [ ] Project scaffolding with clean arch folder structure
- [ ] `go mod init`, install dependencies
- [ ] Domain layer: Task, User, Board entities, Priority VO, domain errors
- [ ] Use case ports: TaskParser, TaskBoard, UserRepository, TaskLogRepository
- [ ] CreateTask use case + unit tests with mocks
- [ ] Rule-based parser gateway
- [ ] Trello gateway + HTTP client
- [ ] SQLite persistence: connection, migrations, user repo, task log repo
- [ ] Composition root wiring all dependencies
- [ ] Telegram bot with polling: `/start`, `/help`, message → card flow
- [ ] Dockerfile (multi-stage build)
- [ ] `.env.example`, Makefile, README

### Phase 2 — Smart Parsing (Week 3)

- [ ] Claude HTTP client
- [ ] Claude parser gateway (implements TaskParser port)
- [ ] Parser chain gateway (LLM → rule-based fallback)
- [ ] Structured prompt engineering for task extraction
- [ ] Label fuzzy matching in Trello gateway
- [ ] Checklist extraction from multi-line messages
- [ ] Unit tests for all parser gateways

### Phase 3 — UX & Commands (Week 4)

- [ ] ListBoards, SelectBoard, SelectList use cases + tests
- [ ] Telegram controller methods for `/boards`, `/setboard`, `/setlist`
- [ ] Inline keyboard builders
- [ ] Callback query routing in Telegram router
- [ ] Confirmation flow: preview → confirm → create
- [ ] Edit flow: tap Edit → modify before creating

### Phase 4 — Production Hardening (Week 5)

- [ ] Webhook mode with TLS
- [ ] Graceful shutdown with signal handling
- [ ] Rate limiting for Telegram + Trello
- [ ] Retry with exponential backoff in HTTP clients
- [ ] Structured logging throughout
- [ ] Health check endpoint
- [ ] Docker Compose with persistent volume
- [ ] `.golangci.yml` with `depguard` for layer enforcement
- [ ] GitHub Actions CI: lint, test, build, docker
- [ ] Integration tests with testcontainers

### Phase 5 — Enhancements (Week 6+)

- [ ] Multi-board routing by keyword/label
- [ ] Voice message → transcription → task
- [ ] Batch task creation from multi-line messages
- [ ] `/list` — show recent tasks
- [ ] Recurring tasks
- [ ] Prometheus metrics
- [ ] K8s deployment manifests

---

## 13. Testing Strategy

| Layer | What | How |
|---|---|---|
| Domain | Entity invariants, value objects | Pure unit tests, zero mocks |
| Use Case | Business orchestration | Unit tests + mocked ports |
| Adapter | Mapping correctness | Unit tests, mock infra |
| Infrastructure | API calls, DB queries | Integration tests, testcontainers |
| E2E | Full message → card flow | Telegram test chat |

---

## 14. Makefile

```makefile
.PHONY: build run test lint mock clean docker-build docker-run

build:
	go build -o bin/bot ./cmd/bot

run:
	go run ./cmd/bot

test:
	go test ./... -v -cover -race

test-unit:
	go test ./internal/domain/... ./internal/usecase/... -v -cover

lint:
	golangci-lint run ./...

mock:
	mockery --all --dir=internal/usecase/port --output=internal/usecase/mocks --outpkg=mocks

docker-build:
	docker build -t telegram-trello-bot -f deployments/Dockerfile .

docker-run:
	docker compose -f deployments/docker-compose.yml up -d

clean:
	rm -rf bin/ coverage.out
```

---

## 15. Deployment

### Dockerfile

```dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o bot ./cmd/bot

FROM alpine:3.19
RUN apk --no-cache add ca-certificates tzdata
RUN adduser -D -g '' appuser
WORKDIR /app
COPY --from=builder /app/bot .
RUN mkdir -p /app/data && chown appuser:appuser /app/data
USER appuser
EXPOSE 8080
CMD ["./bot"]
```

### docker-compose.yml

```yaml
version: "3.8"
services:
  bot:
    build:
      context: ../
      dockerfile: deployments/Dockerfile
    restart: unless-stopped
    env_file: ../.env
    ports:
      - "8080:8080"
    volumes:
      - bot-data:/app/data

volumes:
  bot-data:
```

---

## 16. Risks & Mitigations

| Risk | Mitigation |
|---|---|
| Over-engineering for MVP | Start with 2 use cases, expand incrementally |
| Trello API rate limits | Retry with exponential backoff |
| LLM parsing latency | Parser chain fallback + "processing..." message |
| LLM API costs | Use LLM only when rule-based confidence low |
| Layer violations | `depguard` linter rules + CI enforcement |
| Bot token exposure | `.env`, Docker secrets, never commit tokens |
| SQLite concurrency | WAL mode enabled |
| Message parsing errors | Confirmation step before creating card |

---

## 17. Claude Code Phase Commands

> Each prompt references rules from `CLAUDE.md`. Run prompts in order — each phase builds on the previous.

### Phase 1a: Scaffold

```
Prompt:
Read plan.md Section 4 and CLAUDE.md.

Create the full directory tree and initialize the Go module:

1. Run `go mod init telegram-trello-bot`
2. Create all directories:
   - cmd/bot/
   - internal/domain/entity/
   - internal/domain/valueobject/
   - internal/domain/domainerror/
   - internal/usecase/port/
   - internal/usecase/dto/
   - internal/adapter/controller/
   - internal/adapter/presenter/
   - internal/adapter/gateway/
   - internal/infrastructure/telegram/
   - internal/infrastructure/trello/
   - internal/infrastructure/claude/
   - internal/infrastructure/persistence/
   - internal/infrastructure/config/
   - pkg/httputil/
   - pkg/timeutil/
   - deployments/
3. Create stub files with only `package <name>` declarations.

Rules: NAME-1 (package names), ORG-1 (one type per file), FORBID-2 (no init funcs).

Verify:
  go build ./...        # must compile with zero errors
  find internal/ -name "*.go" | wc -l   # expect ~25 stub files
```

### Phase 1b: Domain Layer

```
Prompt:
Read plan.md Section 5 and CLAUDE.md rules DEP-1, DEP-2, TEST-1, ERR-1.

Implement the domain layer — zero external dependencies:

Files to create/modify:
  internal/domain/entity/task.go           — Task entity with NewTask + options
  internal/domain/entity/user.go           — User entity with NewUser
  internal/domain/entity/board.go          — Board entity
  internal/domain/entity/label.go          — Label value object
  internal/domain/valueobject/priority.go  — Priority enum + NewPriority
  internal/domain/valueobject/task_id.go   — Typed TaskID
  internal/domain/valueobject/telegram_id.go — Typed TelegramID
  internal/domain/domainerror/errors.go    — All sentinel errors

Tests (stdlib only, no testify — TEST-1):
  internal/domain/entity/task_test.go      — TestNewTask_EmptyTitle, TestNewTask_WithOptions, TestTask_IsHighPriority
  internal/domain/valueobject/priority_test.go — TestNewPriority_Valid, TestNewPriority_Invalid

Rules: DEP-2 (domain purity — only stdlib + domain imports), ERR-1 (sentinel errors).

Verify:
  go test ./internal/domain/... -v -cover   # all pass, 0 external deps
  grep -r "import" internal/domain/ | grep -v "_test" | grep -v "domain" | grep -v '"'   # no external imports
```

### Phase 1c: Use Case Layer

```
Prompt:
Read plan.md Section 6 and CLAUDE.md rules DEP-1, DEP-3, PORT-1, PORT-2, PORT-3, TEST-2, DTO-1, DTO-3, FORBID-5.

Implement the use case layer:

Ports (in internal/usecase/port/ — PORT-1):
  task_parser.go       — TaskParser interface: Parse(ctx, message) (*entity.Task, error)
  task_board.go        — TaskBoard interface: CreateCard(ctx, task, boardID, listID) (cardURL string, err error)
                         + ListBoards(ctx, token) ([]entity.Board, error)
                         + ListLists(ctx, token, boardID) ([]string, error)
  user_repository.go   — UserRepository interface: Get(ctx, telegramID) (*entity.User, error)
                         + Save(ctx, user *entity.User) error
  task_log_repository.go — TaskLogRepository interface: Log(ctx, userID, taskTitle, cardURL) error

DTOs (in internal/usecase/dto/ — DTO-1, DTO-3):
  create_task_input.go   — CreateTaskInput{UserID int64, Message string}
  create_task_output.go  — CreateTaskOutput{CardURL, TaskTitle, BoardName, ListName string}

Use Cases (ORG-4 pattern):
  create_task.go         — CreateTaskUseCase: parse → get user → create card → log
  select_board.go        — SelectBoardUseCase: validate board → save preference
  select_list.go         — SelectListUseCase: validate list → save preference
  list_boards.go         — ListBoardsUseCase: fetch boards for user

Tests (testify mocks — TEST-2):
  create_task_test.go    — TestCreateTask_HappyPath, TestCreateTask_ParseFails,
                           TestCreateTask_BoardNotSet, TestCreateTask_CreateCardFails
  select_board_test.go   — TestSelectBoard_HappyPath, TestSelectBoard_InvalidBoard
  list_boards_test.go    — TestListBoards_HappyPath

Rules: FORBID-5 (no SDK types in use cases), PORT-3 (domain types only in ports).

Verify:
  go test ./internal/usecase/... -v -cover   # all pass
  grep -r "telegram\|trello\|claude" internal/usecase/ | grep "import" # must be empty
```

### Phase 1d: Adapter Layer

```
Prompt:
Read plan.md Section 7 and CLAUDE.md rules DEP-1, NAME-4, TEST-3, FORBID-4.

Implement the adapter layer:

Controller:
  internal/adapter/controller/telegram_controller.go
    — TelegramController struct with use case dependencies
    — HandleMessage(ctx, telegramUpdate) method
    — HandleCallback(ctx, callbackQuery) method
    — Maps Telegram input → DTOs → use case Execute → presenter

Presenter:
  internal/adapter/presenter/telegram_presenter.go
    — TelegramPresenter struct
    — FormatTaskCreated(output) string
    — FormatBoardList(boards) string
    — FormatError(err) string
    — Translates domain errors → user-friendly messages (ERR-4)

Gateways (NAME-4 — implementation names reference tech):
  internal/adapter/gateway/trello_gateway.go
    — TrelloGateway implements port.TaskBoard via infra trello.Client
  internal/adapter/gateway/claude_parser_gateway.go
    — ClaudeParserGateway implements port.TaskParser via infra claude.Client
  internal/adapter/gateway/rule_parser_gateway.go
    — RuleParserGateway implements port.TaskParser via regex
  internal/adapter/gateway/parser_chain_gateway.go
    — ParserChainGateway implements port.TaskParser
    — Tries LLM parser first, falls back to rule-based

Tests (TEST-3 — mock infrastructure):
  internal/adapter/gateway/trello_gateway_test.go
  internal/adapter/gateway/parser_chain_gateway_test.go
  internal/adapter/controller/telegram_controller_test.go

Rules: FORBID-4 (no business logic in infra — gateways only translate).

Verify:
  go test ./internal/adapter/... -v -cover   # all pass
```

### Phase 1e: Infrastructure Layer

```
Prompt:
Read plan.md Section 8 and CLAUDE.md rules FORBID-4, FORBID-6, ERR-3, TEST-4.

Implement the infrastructure layer:

Trello HTTP Client:
  internal/infrastructure/trello/client.go   — TrelloClient: HTTP calls to Trello REST API
  internal/infrastructure/trello/models.go   — API response structs (Board, List, Card JSON models)

Claude HTTP Client:
  internal/infrastructure/claude/client.go   — ClaudeClient: HTTP calls to Claude API

SQLite Persistence:
  internal/infrastructure/persistence/sqlite.go             — NewSQLiteDB(path) + RunMigrations
  internal/infrastructure/persistence/user_repo_sqlite.go   — SQLiteUserRepo implements port.UserRepository
  internal/infrastructure/persistence/task_log_repo_sqlite.go — SQLiteTaskLogRepo implements port.TaskLogRepository

Telegram Bot:
  internal/infrastructure/telegram/bot.go      — TelegramBot: init, polling/webhook
  internal/infrastructure/telegram/router.go   — Routes updates → controller methods
  internal/infrastructure/telegram/keyboard.go — Inline keyboard builders

Config:
  internal/infrastructure/config/config.go — Load from .env via viper

Tests (TEST-4 — httptest + in-memory SQLite):
  internal/infrastructure/trello/client_test.go       — httptest for Trello API
  internal/infrastructure/persistence/sqlite_test.go  — ":memory:" SQLite
  internal/infrastructure/claude/client_test.go       — httptest for Claude API

Rules: FORBID-6 (no hardcoded config), ERR-3 (no panics), FORBID-4 (no business logic here).

Verify:
  go test ./internal/infrastructure/... -v -cover   # all pass
```

### Phase 1f: Wire & Compose

```
Prompt:
Read plan.md Section 9 and CLAUDE.md rules ORG-2, FORBID-2, DEP-1.

Implement the composition root and deployment files:

Composition Root:
  cmd/bot/main.go
    — Load config (config.Config)
    — Create infrastructure instances (TrelloClient, ClaudeClient, SQLiteDB, TelegramBot)
    — Create gateways (TrelloGateway, ClaudeParserGateway, RuleParserGateway, ParserChainGateway)
    — Create use cases (CreateTask, SelectBoard, SelectList, ListBoards)
    — Create controller + presenter
    — Create router, wire controller
    — Start bot polling
    — Graceful shutdown with signal handling

Config & Deployment:
  .env.example           — All required env vars with comments
  .gitignore             — Go defaults + .env + bin/ + *.db
  deployments/Dockerfile — Multi-stage: build + scratch/alpine
  deployments/docker-compose.yml — Bot service + volume for SQLite

Rules: ORG-2 (constructor injection only), FORBID-2 (no init() — all setup in main).

Verify:
  go build ./cmd/bot   # must compile
  # Manual: review main.go has NO business logic, only wiring
```

### Phase 1g: Lint & Dependency Verification

```
Prompt:
Read plan.md Sections 10, 14 and CLAUDE.md (all rules).

Set up linting and dependency enforcement:

Linter Config:
  .golangci.yml
    — Enable: govet, errcheck, staticcheck, unused, gosimple, ineffassign, revive
    — Configure depguard with rules matching DEP-1 import matrix:
      * domain packages: deny all except stdlib + domain
      * usecase packages: deny adapter, infrastructure
      * adapter packages: allow all internal
      * infrastructure packages: deny usecase (except port)

Makefile (update/replace):
  — Add phase-specific test targets (see Section 18)
  — Add verify-deps target: grep-based import violation checker
  — Add coverage target: HTML coverage report

Verify Script (belt-and-suspenders with depguard):
  Create scripts/verify-deps.sh:
    — Grep domain/ imports for forbidden packages
    — Grep usecase/ imports for adapter/infrastructure
    — Exit 1 on any violation

Verify:
  make lint              # zero warnings
  make verify-deps       # zero violations
  make test              # all tests pass
  make coverage          # generates coverage.html
```

---

## 18. Makefile Phase Targets

> Add these targets to the project `Makefile` for structured phase execution.

```makefile
.PHONY: build run test lint mock clean docker-build docker-run \
        phase1-scaffold phase1-domain-test phase1-usecase-test \
        phase1-adapter-test phase1-infra-test phase1-wire-verify \
        phase1-test phase2-test phase3-test phase4-test \
        phase-all-test verify-deps coverage

# ── Build & Run ──────────────────────────────────────────────
build:
	go build -o bin/bot ./cmd/bot

run:
	go run ./cmd/bot

# ── Phase 1 Targets ─────────────────────────────────────────
phase1-scaffold:
	@echo "==> Creating project directories..."
	mkdir -p cmd/bot
	mkdir -p internal/domain/{entity,valueobject,domainerror}
	mkdir -p internal/usecase/{port,dto}
	mkdir -p internal/adapter/{controller,presenter,gateway}
	mkdir -p internal/infrastructure/{telegram,trello,claude,persistence,config}
	mkdir -p pkg/{httputil,timeutil}
	mkdir -p deployments
	@echo "==> Scaffold complete."

phase1-domain-test:
	@echo "==> Testing domain layer..."
	go test ./internal/domain/... -v -cover -race

phase1-usecase-test:
	@echo "==> Testing use case layer..."
	go test ./internal/usecase/... -v -cover -race

phase1-adapter-test:
	@echo "==> Testing adapter layer..."
	go test ./internal/adapter/... -v -cover -race

phase1-infra-test:
	@echo "==> Testing infrastructure layer..."
	go test ./internal/infrastructure/... -v -cover -race

phase1-wire-verify:
	@echo "==> Verifying composition root compiles..."
	go build ./cmd/bot

phase1-test: phase1-domain-test phase1-usecase-test phase1-adapter-test phase1-infra-test phase1-wire-verify
	@echo "==> All Phase 1 tests passed."

# ── Phase 2-4 Targets ───────────────────────────────────────
phase2-test:
	@echo "==> Testing Phase 2: parser gateways..."
	go test ./internal/adapter/gateway/... -v -cover -race
	go test ./internal/infrastructure/claude/... -v -cover -race

phase3-test:
	@echo "==> Testing Phase 3: board/list selection + telegram..."
	go test ./internal/usecase/... -v -cover -race -run "Board|List"
	go test ./internal/infrastructure/telegram/... -v -cover -race

phase4-test: lint verify-deps test
	@echo "==> Phase 4 complete: full test + lint."

# ── Aggregated Targets ──────────────────────────────────────
test:
	go test ./... -v -cover -race

test-unit:
	go test ./internal/domain/... ./internal/usecase/... -v -cover

phase-all-test: lint verify-deps test
	@echo "==> All phases passed: lint + verify-deps + test."

lint:
	golangci-lint run ./...

# ── Dependency Verification ─────────────────────────────────
verify-deps:
	@echo "==> Checking import violations (belt-and-suspenders with depguard)..."
	@# Domain must not import usecase, adapter, or infrastructure
	@if grep -rn '"telegram-trello-bot/internal/usecase\|telegram-trello-bot/internal/adapter\|telegram-trello-bot/internal/infrastructure' internal/domain/ 2>/dev/null | grep -v "_test.go"; then \
		echo "FAIL: domain/ has forbidden imports"; exit 1; \
	fi
	@# Use cases must not import adapter or infrastructure
	@if grep -rn '"telegram-trello-bot/internal/adapter\|telegram-trello-bot/internal/infrastructure' internal/usecase/ 2>/dev/null | grep -v "_test.go"; then \
		echo "FAIL: usecase/ has forbidden imports"; exit 1; \
	fi
	@# Infrastructure must not import usecase (except usecase/port)
	@if grep -rn '"telegram-trello-bot/internal/usecase/' internal/infrastructure/ 2>/dev/null | grep -v "_test.go" | grep -v "usecase/port"; then \
		echo "FAIL: infrastructure/ imports usecase (not port)"; exit 1; \
	fi
	@echo "==> All dependency checks passed."

# ── Mocks & Coverage ────────────────────────────────────────
mock:
	mockery --all --dir=internal/usecase/port --output=internal/usecase/mocks --outpkg=mocks

coverage:
	go test ./... -coverprofile=coverage.out -race
	go tool cover -html=coverage.out -o coverage.html
	@echo "==> Coverage report: coverage.html"

# ── Docker ───────────────────────────────────────────────────
docker-build:
	docker build -t telegram-trello-bot -f deployments/Dockerfile .

docker-run:
	docker compose -f deployments/docker-compose.yml up -d

# ── Cleanup ──────────────────────────────────────────────────
clean:
	rm -rf bin/ coverage.out coverage.html
```

---

*Clean Architecture (Uncle Bob) • Go 1.22+ • February 2025*
