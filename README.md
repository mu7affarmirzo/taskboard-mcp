# Telegram-Trello Bot

A Telegram bot that converts natural language messages into Trello cards. Built with Go using Clean Architecture principles.

Send a message like _"Fix the login bug by Friday #backend urgent"_ and the bot will parse it, show a preview, and create a Trello card on confirmation.

## Architecture

```
cmd/bot/main.go          ← Composition Root
internal/
  domain/                ← Entities, Value Objects, Errors (zero external deps)
  usecase/               ← Business logic, ports (interfaces), DTOs
  adapter/               ← Controllers, Presenters, Gateways
  infrastructure/        ← Telegram, Trello, Claude, SQLite, Config
pkg/                     ← Shared utilities (httputil, timeutil, ratelimit)
```

Dependency flow: `domain` ← `usecase` ← `adapter` ← `infrastructure` ← `cmd/bot`

## Quick Start

### Prerequisites

- Go 1.24+
- Telegram bot token (from [@BotFather](https://t.me/BotFather))
- Trello API key (from [trello.com/app-key](https://trello.com/app-key))
- (Optional) Claude API key for AI-powered task parsing

### Setup

```bash
cp .env.example .env
# Edit .env with your tokens

make run
```

### Environment Variables

| Variable | Required | Description |
|---|---|---|
| `TELEGRAM_BOT_TOKEN` | Yes | Bot token from BotFather |
| `TRELLO_API_KEY` | Yes | Trello API key |
| `CLAUDE_API_KEY` | No | Anthropic API key for AI parsing |
| `CLAUDE_MODEL` | No | Claude model ID |
| `TELEGRAM_MODE` | No | `polling` (default) or `webhook` |
| `TELEGRAM_WEBHOOK_URL` | No | Public HTTPS URL for webhook mode |
| `DATABASE_PATH` | No | SQLite path (default: `./data/bot.db`) |
| `PORT` | No | Health check port (default: `8080`) |
| `LOG_LEVEL` | No | `debug`, `info`, `warn`, `error` |

## Docker Deployment

```bash
# Build and run
make deploy

# View logs
make logs

# Stop
make stop
```

Or manually:

```bash
docker compose -f deployments/docker-compose.yml up -d --build
```

## Bot Commands

| Command | Description |
|---|---|
| `/start` | Welcome message |
| `/help` | Show usage tips and commands |
| `/boards` | List your Trello boards and select a default |

### Task Creation Flow

1. Send any text message describing a task
2. Bot parses it and shows a preview with title, priority, due date, labels
3. Tap **Create**, **Edit**, or **Cancel**
4. On confirm, the card is created in your default Trello board/list

### Parsing Tips

- Include "urgent" or "high priority" to set high priority
- Add `#label` to tag your task
- Say "due Friday" or "by March 15" to set a deadline

## Development

```bash
# Run all tests
make test

# Run linter
make lint

# Verify clean architecture dependency rules
make verify-deps

# Generate coverage report
make coverage
```

## Project Structure

```
.
├── cmd/bot/main.go
├── internal/
│   ├── domain/
│   │   ├── entity/          # Task, User, Label
│   │   ├── valueobject/     # Priority, TelegramID
│   │   └── domainerror/     # Sentinel errors
│   ├── usecase/
│   │   ├── port/            # Interfaces (TaskParser, TaskBoard, etc.)
│   │   ├── dto/             # Input/Output structs
│   │   ├── create_task.go
│   │   ├── parse_task.go
│   │   ├── confirm_task.go
│   │   ├── list_boards.go
│   │   ├── list_lists.go
│   │   ├── select_board.go
│   │   └── select_list.go
│   ├── adapter/
│   │   ├── controller/      # TelegramController
│   │   ├── presenter/       # TelegramPresenter
│   │   └── gateway/         # TrelloGateway, ClaudeParserGateway, etc.
│   └── infrastructure/
│       ├── telegram/        # Bot, Router, Keyboard builders
│       ├── trello/          # HTTP client for Trello API
│       ├── claude/          # HTTP client for Claude API
│       ├── persistence/     # SQLite repos
│       ├── config/          # Viper-based config loader
│       ├── state/           # In-memory pending task store
│       └── health/          # /healthz endpoint
├── pkg/
│   ├── httputil/            # Retry HTTP client with backoff
│   ├── timeutil/            # Natural date parsing
│   └── ratelimit/           # Token-bucket rate limiter
├── deployments/
│   ├── Dockerfile
│   └── docker-compose.yml
└── Makefile
```

## License

MIT
