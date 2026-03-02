# Clean Architecture Rules — Telegram-Trello Bot (Go)

> Machine-referenced rules for Claude Code prompts. Each rule has a short ID (e.g., DEP-1) used in `plan.md` Section 17 phase commands.

---

## 1. Dependency Rules (DEP)

### DEP-1: Layer Import Matrix

| Importing Layer | domain | usecase | usecase/port | adapter | infrastructure | pkg |
|-----------------|--------|---------|--------------|---------|----------------|-----|
| **domain** | self | NO | NO | NO | NO | NO |
| **usecase** | YES | self | self | NO | NO | NO |
| **adapter** | YES | YES | YES | self | YES | YES |
| **infrastructure** | YES | NO | YES | YES | self | YES |
| **cmd/bot (main)** | YES | YES | YES | YES | YES | YES |

### DEP-2: Domain Purity

Domain packages (`internal/domain/**`) must have **zero** imports outside of:
- Go standard library (`fmt`, `errors`, `time`, `strings`, etc.)
- Other domain packages (`internal/domain/*`)

No third-party libraries. No `usecase`, `adapter`, or `infrastructure` imports.

### DEP-3: Port Ownership

Interfaces (ports) live in `internal/usecase/port/`. Use cases depend on these ports. Adapters and infrastructure implement them. Ports never import their implementations.

### DEP-4: No Circular Imports

No package may import another package that directly or transitively imports it back. If you detect a cycle, extract shared types into `domain` or create a new `port` interface.

---

## 2. Naming Conventions (NAME)

### NAME-1: Package Names

- All lowercase, single word when possible: `entity`, `valueobject`, `port`, `dto`
- Multi-word uses no separator: `domainerror`, `httputil`, `timeutil`
- Package name must match directory name exactly

### NAME-2: File Names

- `snake_case.go` for all Go source files
- Test files: `<name>_test.go` in same package
- One primary type per file; file named after the type: `task.go` for `Task`, `priority.go` for `Priority`

### NAME-3: Interface Names

- No `I` prefix. Use descriptive noun/verb names: `TaskParser`, `TaskBoard`, `UserRepository`
- Port interfaces describe **what** the use case needs, not **how** it's implemented

### NAME-4: Struct Names

- Implementation structs reference their tech: `TrelloGateway`, `SQLiteUserRepo`, `ClaudeParserGateway`
- Adapter structs: `TelegramController`, `TelegramPresenter`

### NAME-5: Constructors

- `New<Type>(deps...) *Type` pattern: `NewCreateTaskUseCase(parser TaskParser, board TaskBoard) *CreateTaskUseCase`
- Constructor receives all dependencies as parameters (see ORG-2)

### NAME-6: Receivers

- Short, 1-2 letter receivers: `func (uc *CreateTaskUseCase) Execute(...)`, `func (t *Task) Title()`
- Consistent receiver name per type across all methods

---

## 3. Testing Rules (TEST)

### TEST-1: Domain Tests — Pure stdlib

Domain tests use only `testing` from stdlib. No testify, no mocks. Test entity invariants and value object validation.

```
internal/domain/entity/task_test.go      → TestNewTask_EmptyTitle, TestNewTask_WithOptions
internal/domain/valueobject/priority_test.go → TestNewPriority_Valid, TestNewPriority_Invalid
```

### TEST-2: Use Case Tests — Testify Mocks

Use case tests mock all ports using `testify/mock`. Test business logic paths: happy path, validation errors, port failures.

```
internal/usecase/create_task_test.go     → mock TaskParser, TaskBoard, UserRepository
internal/usecase/select_board_test.go    → mock UserRepository, TaskBoard
```

### TEST-3: Adapter Tests — Mock Infrastructure

Adapter tests mock the infrastructure clients they wrap. Test mapping/translation logic.

```
internal/adapter/gateway/trello_gateway_test.go → mock trello.Client
internal/adapter/gateway/parser_chain_gateway_test.go → mock both parsers
```

### TEST-4: Infrastructure Tests — httptest & In-Memory

Infrastructure tests use `httptest` for HTTP clients and in-memory SQLite for persistence. No real external services.

```
internal/infrastructure/trello/client_test.go     → httptest server
internal/infrastructure/persistence/sqlite_test.go → ":memory:" database
```

### TEST-5: Test File Placement

Test files live next to the code they test, in the same package. No separate `test/` directory.

### TEST-6: Test Naming

`Test<Function>_<Scenario>` pattern: `TestNewTask_EmptyTitle_ReturnsError`, `TestCreateTask_HappyPath`.

---

## 4. Code Organization (ORG)

### ORG-1: One Type Per File

Each exported struct/interface gets its own file. Helpers and unexported types can colocate with their primary type.

### ORG-2: Constructor Injection Only

All dependencies injected via constructor parameters. No `init()`, no global registries, no service locators.

```go
// YES
func NewCreateTaskUseCase(parser port.TaskParser, board port.TaskBoard) *CreateTaskUseCase {
    return &CreateTaskUseCase{parser: parser, board: board}
}

// NO
var globalParser port.TaskParser  // FORBID-1
func init() { /* ... */ }        // FORBID-2
```

### ORG-3: Context Propagation

All use case `Execute` methods and port methods accept `context.Context` as first parameter.

```go
func (uc *CreateTaskUseCase) Execute(ctx context.Context, input dto.CreateTaskInput) (*dto.CreateTaskOutput, error)
```

### ORG-4: Use Case Structure

Every use case follows this pattern:
1. Struct with port dependencies as private fields
2. `New<Name>UseCase(ports...) *<Name>UseCase` constructor
3. Single `Execute(ctx, input) (output, error)` method

---

## 5. Error Handling (ERR)

### ERR-1: Sentinel Domain Errors

All domain errors are package-level `var` sentinels in `internal/domain/domainerror/errors.go`, created with `errors.New()`.

```go
var ErrEmptyTaskTitle = errors.New("task title cannot be empty")
```

### ERR-2: Wrapping with %w

Non-domain layers wrap errors with `fmt.Errorf("context: %w", err)` to preserve the error chain. Always add context about what operation failed.

### ERR-3: No Panics in Business Logic

Never `panic()` in domain, usecase, or adapter code. Return errors. Only `main.go` may `log.Fatal` on startup failures.

### ERR-4: Error Type Boundaries

Domain errors cross layer boundaries via `errors.Is()` / `errors.As()`. Adapters translate domain errors into user-facing messages. Infrastructure errors are wrapped before returning through ports.

---

## 6. Port/Interface Rules (PORT)

### PORT-1: Interfaces in usecase/port/

All port interfaces live in `internal/usecase/port/`. One file per port. The port package imports only `domain` and stdlib.

### PORT-2: Interface Segregation

Keep interfaces small and focused. A port should have 1-3 methods. Split large interfaces into multiple ports.

```go
// YES — focused
type TaskParser interface {
    Parse(ctx context.Context, message string) (*entity.Task, error)
}

// NO — too broad
type Repository interface {
    GetUser(...) ...
    SaveUser(...) ...
    GetTask(...) ...
    SaveTask(...) ...
}
```

### PORT-3: Method Signatures

Port methods accept/return domain types and DTOs, never infrastructure types. Parameters: `(ctx context.Context, domainTypes...)`. Returns: `(domainTypes/DTOs, error)`.

### PORT-4: No Implementation Awareness

Ports must not reference or be aware of their implementations. No `trello`, `sqlite`, or `claude` types in port signatures.

---

## 7. DTO Rules (DTO)

### DTO-1: DTOs in usecase/dto/

All DTOs live in `internal/usecase/dto/`. Used for use case input/output boundaries.

### DTO-2: When DTOs vs Entities

- **Input DTOs**: raw data from controllers → use cases (before domain validation)
- **Output DTOs**: use case results → presenters (after business logic)
- **Entities**: within domain and use case logic (fully validated)

### DTO-3: Plain Structs Only

DTOs are plain structs with exported fields. No methods, no validation logic, no interfaces. Validation happens in use cases or entity constructors.

```go
type CreateTaskInput struct {
    UserID    int64
    Message   string
}

type CreateTaskOutput struct {
    CardURL   string
    TaskTitle string
    BoardName string
    ListName  string
}
```

---

## 8. Forbidden Patterns (FORBID)

### FORBID-1: No Global State

No package-level mutable variables (except sentinel errors). No singletons. All state lives in structs created by constructors.

### FORBID-2: No init() Functions

No `init()` in any package. All initialization happens explicitly in `cmd/bot/main.go`.

### FORBID-3: No God Structs

No struct with more than 5 dependencies. If a struct needs more, split the responsibility into multiple use cases or introduce a facade.

### FORBID-4: No Business Logic in Infrastructure

Infrastructure packages contain only client wrappers and data access. Conditional logic, branching on business rules, and validation belong in domain or use case layers.

### FORBID-5: No SDK Types in Use Cases

Use case layer must not import Telegram SDK, Trello API structs, Claude client types, or any infrastructure type. Use ports and domain types only.

### FORBID-6: No Hardcoded Config

No hardcoded URLs, tokens, timeouts, or magic numbers. All config comes from environment variables loaded via `config.go`.

### FORBID-7: No Naked Goroutines in Business Logic

No `go func()` in domain or use case layers. Concurrency is managed by infrastructure (Telegram polling) or composition root.

---

## Quick Reference: Layer Import Table

```
domain/         → stdlib only
usecase/        → domain/ + stdlib
usecase/port/   → domain/ + stdlib
usecase/dto/    → stdlib only (plain structs)
adapter/        → usecase/ + domain/ + infrastructure/ + pkg/ + stdlib
infrastructure/ → usecase/port/ + domain/ + adapter/ (router→controller) + pkg/ + stdlib
cmd/bot/        → everything
pkg/            → stdlib only (shared utilities)
```

---

## Pre-Commit Checklist

Before committing, verify:

- [ ] **DEP-1**: No import violations — run `make verify-deps`
- [ ] **DEP-2**: Domain has zero external imports
- [ ] **TEST-1/2/3/4**: All tests pass — run `make test`
- [ ] **FORBID-1**: No global mutable state (grep for `var ` in non-test files)
- [ ] **FORBID-2**: No `init()` functions (grep for `func init()`)
- [ ] **ERR-3**: No `panic()` outside tests
- [ ] **NAME-1**: Package names match directory names
- [ ] **ORG-2**: All dependencies injected via constructors
- [ ] **PORT-4**: Ports contain no infrastructure types
