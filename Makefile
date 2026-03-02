.PHONY: build run test lint mock clean docker-build docker-run \
        phase1-scaffold phase1-domain-test phase1-usecase-test \
        phase1-adapter-test phase1-infra-test phase1-wire-verify \
        phase1-test phase2-test phase3-test phase4-test \
        phase-all-test verify-deps coverage deploy logs stop

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

# ── Deploy ──────────────────────────────────────────────────
deploy:
	docker compose -f deployments/docker-compose.yml up -d --build

logs:
	docker compose -f deployments/docker-compose.yml logs -f

stop:
	docker compose -f deployments/docker-compose.yml down

# ── Cleanup ──────────────────────────────────────────────────
clean:
	rm -rf bin/ coverage.out coverage.html
