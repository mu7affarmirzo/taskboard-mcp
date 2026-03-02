# Implementation Plan — Missing Features & Gaps

> Comprehensive plan to finish all missing items from `plan.md` Phases 3-5, fill quality gaps, and get the bot running in Docker.

---

## Current State

- **Compiles**: `go build ./...` passes
- **Tests**: ~116 tests, all passing
- **Layers built**: Domain, Use Case, Adapter, Infrastructure, Composition Root
- **Core flow works**: message → parse → create Trello card (no confirmation step)
- **Board/List selection**: works via inline keyboards + callbacks
- **Docker files exist**: Dockerfile + docker-compose.yml (not yet validated end-to-end)

---

## Phase A — Confirmation Flow (Preview → Confirm → Create)

**Goal**: When a user sends a message, the bot parses it, shows a preview with a confirm keyboard, and only creates the Trello card after the user taps "Create".

### A1. Add ParseTask use case (parse-only, no card creation)

**Why**: The current `CreateTaskUseCase.Execute` does parse + create in one shot. We need a two-step flow: first parse and return a preview, then create on confirmation.

**Files to create**:
- `internal/usecase/dto/parse_task_output.go`

```go
type ParseTaskOutput struct {
    TaskTitle   string
    Description string
    DueDate     *time.Time
    Priority    string
    Labels      []string
    Checklist   []string
    BoardID     string   // user's default board
    ListID      string   // user's default list
}
```

- `internal/usecase/parse_task.go` — `ParseTaskUseCase`
  - Find user → validate board/list configured
  - Parse message via `TaskParser` port
  - Return `ParseTaskOutput` (no card creation, no logging)

**Files to create (tests)**:
- `internal/usecase/parse_task_test.go` — happy path, parse fails, board not set

**Rules**: ORG-4, TEST-2, FORBID-5

---

### A2. Add ConfirmTask use case (create card from pre-parsed data)

**Why**: After user confirms, we need to create the card from the data we already parsed (no re-parsing).

**Files to create**:
- `internal/usecase/dto/confirm_task_input.go`

```go
type ConfirmTaskInput struct {
    TelegramID  int64
    Title       string
    Description string
    DueDate     *time.Time
    Priority    string
    Labels      []string
}
```

- `internal/usecase/confirm_task.go` — `ConfirmTaskUseCase`
  - Find user → resolve label IDs → build `CreateCardParams` → create card → log → return `CreateTaskOutput`

**Files to create (tests)**:
- `internal/usecase/confirm_task_test.go`

**Rules**: ORG-4, TEST-2

---

### A3. Add in-memory pending task store

**Why**: Between parse (preview) and confirm (create), the bot must hold the parsed task data per user. This is conversational state, not domain persistence.

**Approach**: Simple `sync.Map` based store in infrastructure layer, keyed by `telegramID`.

**Files to create**:
- `internal/infrastructure/state/pending_store.go`

```go
type PendingTask struct {
    Title       string
    Description string
    DueDate     *time.Time
    Priority    string
    Labels      []string
    Checklist   []string
    CreatedAt   time.Time  // for TTL-based cleanup
}

type PendingStore struct { ... }
func NewPendingStore() *PendingStore
func (s *PendingStore) Set(telegramID int64, task PendingTask)
func (s *PendingStore) Get(telegramID int64) (PendingTask, bool)
func (s *PendingStore) Delete(telegramID int64)
```

- `internal/infrastructure/state/pending_store_test.go`

**Rules**: FORBID-1 (no globals — store is a struct injected via constructor), TEST-4

---

### A4. Update Router to use two-step flow

**Files to modify**:
- `internal/adapter/controller/telegram_controller.go`
  - Add `HandleParseTask(ctx, telegramID, text)` method → calls `ParseTaskUseCase`
  - Add `HandleConfirmTask(ctx, telegramID)` method → reads from `PendingStore` → calls `ConfirmTaskUseCase`
  - Add `HandleCancelTask(ctx, telegramID)` method → deletes from `PendingStore`
- `internal/adapter/presenter/telegram_presenter.go`
  - `FormatTaskPreview` already exists — verify it matches `ParseTaskOutput` fields
- `internal/infrastructure/telegram/router.go`
  - Plain message → `HandleParseTask` → send preview with `BuildConfirmKeyboard()`
  - `confirm:create` → `HandleConfirmTask` → send `FormatTaskCreated`
  - `confirm:cancel` → `HandleCancelTask` → send "Task cancelled"
  - `confirm:edit` → `HandleCancelTask` + send "Send your edited message"

**Files to update (tests)**:
- `internal/adapter/controller/telegram_controller_test.go`
- `internal/infrastructure/telegram/router_test.go`

**Rules**: FORBID-4 (no business logic in router — router just dispatches)

---

### A5. Wire new components in main.go

**Files to modify**:
- `cmd/bot/main.go`
  - Create `PendingStore`
  - Create `ParseTaskUseCase`, `ConfirmTaskUseCase`
  - Pass them to `TelegramController`

---

## Phase B — Production Hardening

### B1. HTTP retry client with exponential backoff (`pkg/httputil`)

**Why**: Trello and Claude API calls can transiently fail. The `pkg/httputil/client.go` file is currently a stub.

**Files to modify**:
- `pkg/httputil/client.go`

```go
type RetryConfig struct {
    MaxRetries int
    BaseDelay  time.Duration
    MaxDelay   time.Duration
}

type RetryClient struct {
    client *http.Client
    config RetryConfig
}

func NewRetryClient(cfg RetryConfig) *RetryClient
func (rc *RetryClient) Do(req *http.Request) (*http.Response, error)
```

- Retry on 429, 500, 502, 503, 504 status codes
- Exponential backoff with jitter
- Respect `Retry-After` header

**Files to create (tests)**:
- `pkg/httputil/client_test.go` — httptest with flaky server

**Then update**:
- `internal/infrastructure/trello/client.go` — use `RetryClient` instead of plain `http.Client`
- `internal/infrastructure/claude/client.go` — use `RetryClient` instead of plain `http.Client`
- Constructor injection: `NewClient(apiKey string, httpClient *http.Client)` so retry client is passed in from main.go

**Rules**: PORT-4 (pkg has no business logic), FORBID-6 (retry config from env)

---

### B2. Rate limiter

**Why**: Telegram enforces ~30 msg/sec per bot, Trello has 100 req/10sec per token.

**Files to create**:
- `pkg/ratelimit/limiter.go` — token-bucket or sliding-window rate limiter using `golang.org/x/time/rate`

```go
type Limiter struct {
    limiter *rate.Limiter
}

func NewLimiter(rps float64, burst int) *Limiter
func (l *Limiter) Wait(ctx context.Context) error
```

- `pkg/ratelimit/limiter_test.go`

**Integration**:
- Wrap Trello client calls with limiter in `TrelloGateway`
- Wrap Telegram `Send` calls with limiter in `Router`

**Rules**: FORBID-1 (limiter is a struct, not global)

---

### B3. Webhook mode

**Why**: Polling is fine for dev, but production Telegram bots should use webhooks.

**Files to modify**:
- `internal/infrastructure/telegram/bot.go`
  - Add `StartWebhook(addr, certFile, keyFile string)` method
  - Uses `tgbotapi.NewWebhook()` + `tgbotapi.ListenForWebhook()`
  - Route updates through the same `Router`
- `cmd/bot/main.go`
  - Check `cfg.TelegramMode`: if `"webhook"` → `bot.StartWebhook(...)`, else → `bot.StartPolling()`

**Files to modify (tests)**:
- `internal/infrastructure/telegram/bot_test.go` (new) — verify webhook setup with httptest

**Config additions to `.env.example`**:
```
TELEGRAM_WEBHOOK_URL=https://yourdomain.com/webhook
TELEGRAM_WEBHOOK_CERT=    # optional: path to self-signed cert
TELEGRAM_WEBHOOK_KEY=     # optional: path to key file
```

---

### B4. Health check endpoint

**Why**: Docker and orchestrators need a health probe.

**Files to create**:
- `internal/infrastructure/health/handler.go`

```go
func NewHealthHandler(db *sql.DB) http.Handler
// GET /healthz → 200 {"status":"ok"} if DB ping succeeds, 503 otherwise
```

- `internal/infrastructure/health/handler_test.go`

**Integration**:
- `cmd/bot/main.go` — start a background `http.ListenAndServe(":"+cfg.Port, mux)` with `/healthz`
- `deployments/Dockerfile` — add `HEALTHCHECK CMD wget -qO- http://localhost:8080/healthz || exit 1`
- `deployments/docker-compose.yml` — add `healthcheck` block

---

### B5. Structured logging improvements

**Why**: Current logging is basic. Need request-scoped context with user/chat IDs.

**Files to modify**:
- `internal/infrastructure/telegram/router.go`
  - Add `slog.With("user_id", userID, "chat_id", chatID)` to each handler path
  - Log at start/end of each route with duration
- Propagate logger through context where useful

---

### B6. Graceful shutdown improvements

**Files to modify**:
- `cmd/bot/main.go`
  - Close DB after bot stops (current `defer db.Close()` is fine)
  - If webhook mode: shut down HTTP server with `server.Shutdown(ctx)`
  - Add shutdown timeout (e.g., 10s)

---

## Phase C — CI/CD & Linting

### C1. GitHub Actions workflow

**Files to create**:
- `.github/workflows/ci.yml`

```yaml
name: CI
on: [push, pull_request]
jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.22' }
      - uses: golangci/golangci-lint-action@v6
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.22' }
      - run: make test
      - run: make verify-deps
  build:
    runs-on: ubuntu-latest
    needs: [lint, test]
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.22' }
      - run: make build
  docker:
    runs-on: ubuntu-latest
    needs: [build]
    steps:
      - uses: actions/checkout@v4
      - run: make docker-build
```

---

### C2. Fix Dockerfile Go version

**Why**: `go.mod` says `go 1.25.4` but Dockerfile uses `golang:1.22-alpine`. These must match.

**Files to modify**:
- `deployments/Dockerfile` — change base image to match go.mod version, or pin go.mod to a released version
- Recommendation: align both to `go 1.22` (currently released) since `1.25.4` does not exist as a Docker image. Update `go.mod` to `go 1.22`.

---

## Phase D — Missing Tests & Quality

### D1. Add `label_test.go`

**Why**: Only entity without tests.

**Files to create**:
- `internal/domain/entity/label_test.go`
  - `TestNewLabel_Valid`
  - `TestLabel_Getters`

**Rules**: TEST-1 (stdlib only)

---

### D2. Add `config_test.go`

**Files to create**:
- `internal/infrastructure/config/config_test.go`
  - Test that `Load()` reads env vars correctly
  - Use `t.Setenv()` for isolated tests

---

### D3. Improve presenter `FormatBoardSelected` / `FormatListSelected`

**Why**: Currently show raw IDs (e.g., `"Board set to *abc123def*"`). Should show board/list names.

**Files to modify**:
- `internal/adapter/presenter/telegram_presenter.go`
  - `FormatBoardSelected(boardName string)` — accept name, not ID
  - `FormatListSelected(listName string)` — accept name, not ID
- `internal/infrastructure/telegram/router.go` — pass names from callback flow

---

## Phase E — Extract `pkg/timeutil`

**Why**: Natural language date parsing lives inline in `rule_parser_gateway.go`. The plan calls for it in `pkg/timeutil`.

### E1. Extract `parseNaturalDate` to `pkg/timeutil`

**Files to modify**:
- `pkg/timeutil/parse.go`
  - Move `parseNaturalDate` from `rule_parser_gateway.go` to `ParseNaturalDate` (exported)
  - Add more format support if needed

- `internal/adapter/gateway/rule_parser_gateway.go`
  - Import `pkg/timeutil` and call `timeutil.ParseNaturalDate()`

**Files to create**:
- `pkg/timeutil/parse_test.go`
  - `TestParseNaturalDate_Today`, `_Tomorrow`, `_Weekday`, `_ISO`, `_MonthDay`

**Rules**: pkg/ imports only stdlib

---

## Phase F — README & Documentation

### F1. Create `README.md`

**File to create**: `README.md`

Content outline:
- Project description (Telegram bot → Trello cards via Clean Architecture)
- Architecture diagram (from plan.md Section 2)
- Quick start (prerequisites, `.env` setup, `make run`)
- Docker deployment (`make docker-build && make docker-run`)
- Commands reference (`/start`, `/help`, `/boards`)
- Development (`make test`, `make lint`, `make verify-deps`)
- Project structure (condensed tree)
- License

---

## Phase G — Docker Deployment (Run the Bot)

### G1. Fix Docker image compatibility

**Files to modify**:
- `deployments/Dockerfile`
  - Fix Go version to match `go.mod`
  - Add `HEALTHCHECK` instruction
  - Set `DATABASE_PATH` default to `/app/data/bot.db`

Updated Dockerfile:

```dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o bot ./cmd/bot

FROM alpine:3.19
RUN apk --no-cache add ca-certificates tzdata wget
RUN adduser -D -g '' appuser
WORKDIR /app
COPY --from=builder /app/bot .
RUN mkdir -p /app/data && chown appuser:appuser /app/data
USER appuser
ENV DATABASE_PATH=/app/data/bot.db
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=5s --retries=3 \
  CMD wget -qO- http://localhost:8080/healthz || exit 1
CMD ["./bot"]
```

---

### G2. Update docker-compose.yml

**Files to modify**:
- `deployments/docker-compose.yml`

```yaml
version: "3.8"
services:
  bot:
    build:
      context: ../
      dockerfile: deployments/Dockerfile
    container_name: telegram-trello-bot
    restart: unless-stopped
    env_file: ../.env
    ports:
      - "${PORT:-8080}:8080"
    volumes:
      - bot-data:/app/data
    healthcheck:
      test: ["CMD", "wget", "-qO-", "http://localhost:8080/healthz"]
      interval: 30s
      timeout: 5s
      retries: 3
      start_period: 10s
    logging:
      driver: json-file
      options:
        max-size: "10m"
        max-file: "3"

volumes:
  bot-data:
```

---

### G3. Add `make deploy` convenience target

**Files to modify**:
- `Makefile` — add:

```makefile
deploy:
	docker compose -f deployments/docker-compose.yml up -d --build

logs:
	docker compose -f deployments/docker-compose.yml logs -f

stop:
	docker compose -f deployments/docker-compose.yml down
```

---

### G4. Validate end-to-end Docker run

**Manual checklist**:
1. Copy `.env.example` → `.env`, fill in real tokens
2. `make deploy`
3. `docker compose -f deployments/docker-compose.yml ps` → bot is `Up (healthy)`
4. Send `/start` to bot on Telegram → get welcome message
5. Send `/boards` → see inline keyboard
6. Select board → select list → send a message → see preview → confirm → card created
7. `make logs` → structured JSON logs visible
8. `make stop` → clean shutdown

---

## Implementation Order

Execute phases in this order, as each builds on the previous:

| Step | Phase | What | Estimated Effort |
|------|-------|------|-----------------|
| 1 | C2 | Fix `go.mod` Go version to match Docker | 5 min |
| 2 | D1-D2 | Add missing tests (label, config) | 30 min |
| 3 | E1 | Extract `pkg/timeutil` | 20 min |
| 4 | A1-A2 | ParseTask + ConfirmTask use cases | 1 hr |
| 5 | A3 | PendingStore (in-memory state) | 30 min |
| 6 | A4 | Update controller/router for confirm flow | 1 hr |
| 7 | A5 | Wire in main.go | 15 min |
| 8 | D3 | Fix presenter board/list name display | 15 min |
| 9 | B1 | HTTP retry client (`pkg/httputil`) | 45 min |
| 10 | B2 | Rate limiter (`pkg/ratelimit`) | 30 min |
| 11 | B4 | Health check endpoint | 30 min |
| 12 | B3 | Webhook mode | 45 min |
| 13 | B5-B6 | Logging + graceful shutdown improvements | 30 min |
| 14 | G1-G3 | Docker fixes + deploy targets | 30 min |
| 15 | C1 | GitHub Actions CI | 20 min |
| 16 | F1 | README.md | 30 min |
| 17 | G4 | End-to-end Docker validation | 15 min |

**Total estimated effort: ~8 hours**

---

## Files Summary

### New files to create (17)

```
internal/usecase/parse_task.go
internal/usecase/parse_task_test.go
internal/usecase/confirm_task.go
internal/usecase/confirm_task_test.go
internal/usecase/dto/parse_task_output.go
internal/usecase/dto/confirm_task_input.go
internal/infrastructure/state/pending_store.go
internal/infrastructure/state/pending_store_test.go
internal/infrastructure/health/handler.go
internal/infrastructure/health/handler_test.go
internal/infrastructure/config/config_test.go
internal/domain/entity/label_test.go
pkg/httputil/client_test.go
pkg/ratelimit/limiter.go
pkg/ratelimit/limiter_test.go
pkg/timeutil/parse_test.go
.github/workflows/ci.yml
README.md
```

### Existing files to modify (14)

```
go.mod                                              (Go version)
cmd/bot/main.go                                     (wire new components, health server, webhook mode)
pkg/httputil/client.go                              (implement retry client)
pkg/timeutil/parse.go                               (extract date parsing)
internal/adapter/controller/telegram_controller.go  (add parse/confirm/cancel handlers)
internal/adapter/controller/telegram_controller_test.go
internal/adapter/presenter/telegram_presenter.go    (fix board/list name display)
internal/adapter/presenter/telegram_presenter_test.go
internal/adapter/gateway/rule_parser_gateway.go     (use pkg/timeutil)
internal/infrastructure/telegram/router.go          (two-step confirm flow)
internal/infrastructure/telegram/router_test.go
internal/infrastructure/telegram/bot.go             (add webhook mode)
internal/infrastructure/trello/client.go            (accept http.Client for retry)
internal/infrastructure/claude/client.go            (accept http.Client for retry)
deployments/Dockerfile
deployments/docker-compose.yml
Makefile
.env.example
```

---

## Architecture Compliance Checklist

Every change must satisfy:

- [ ] **DEP-1/DEP-2**: No import violations (domain stays pure)
- [ ] **PORT-1/PORT-3**: New ports in `usecase/port/`, domain types only
- [ ] **ORG-2**: All new dependencies injected via constructors
- [ ] **ORG-4**: New use cases follow Execute pattern
- [ ] **FORBID-1**: PendingStore is a struct, not a global
- [ ] **FORBID-2**: No `init()` functions
- [ ] **FORBID-5**: No SDK types in use case layer
- [ ] **TEST-1**: Domain tests use stdlib only
- [ ] **TEST-2**: Use case tests use testify mocks
- [ ] **TEST-4**: Infrastructure tests use httptest / in-memory
- [ ] **ERR-1/ERR-2**: Sentinel errors + `%w` wrapping
- [ ] `make test` passes
- [ ] `make lint` passes
- [ ] `make verify-deps` passes
