package main

import (
	"context"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"telegram-trello-bot/internal/adapter/controller"
	"telegram-trello-bot/internal/adapter/gateway"
	"telegram-trello-bot/internal/adapter/presenter"
	"telegram-trello-bot/internal/infrastructure/claude"
	"telegram-trello-bot/internal/infrastructure/config"
	"telegram-trello-bot/internal/infrastructure/health"
	"telegram-trello-bot/internal/infrastructure/miniapp"
	"telegram-trello-bot/internal/infrastructure/persistence"
	"telegram-trello-bot/internal/infrastructure/state"
	"telegram-trello-bot/internal/infrastructure/telegram"
	infraTrello "telegram-trello-bot/internal/infrastructure/trello"
	"telegram-trello-bot/internal/usecase"
	"telegram-trello-bot/pkg/httputil"
	"telegram-trello-bot/web"
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
	defer func() { _ = db.Close() }()

	httpClient := httputil.NewRetryHTTPClient(httputil.RetryConfig{
		MaxRetries: 3,
		BaseDelay:  500 * time.Millisecond,
		MaxDelay:   5 * time.Second,
	}, 30*time.Second)

	trelloClient := infraTrello.NewClient(cfg.TrelloAPIKey, httpClient)
	claudeClient := claude.NewClient(cfg.ClaudeAPIKey, cfg.ClaudeModel, httpClient)

	// Repositories
	userRepo := persistence.NewUserRepoSQLite(db)
	taskLogRepo := persistence.NewTaskLogRepoSQLite(db)

	// Gateways
	trelloGw := gateway.NewTrelloGateway(trelloClient)
	llmParser := gateway.NewClaudeParserGateway(claudeClient)
	ruleParser := gateway.NewRuleParserGateway()
	parserChain := gateway.NewParserChainGateway(llmParser, ruleParser, logger)

	intentParser := gateway.NewClaudeIntentGateway(claudeClient)
	intentChain := gateway.NewIntentChainGateway(intentParser, parserChain, logger)

	// State
	pendingStore := state.NewPendingStore()

	// Use Cases
	createTaskUC := usecase.NewCreateTaskUseCase(parserChain, trelloGw, trelloGw, userRepo, taskLogRepo)
	parseTaskUC := usecase.NewParseTaskUseCase(parserChain, userRepo)
	confirmTaskUC := usecase.NewConfirmTaskUseCase(trelloGw, trelloGw, userRepo, taskLogRepo)
	listBoardsUC := usecase.NewListBoardsUseCase(trelloGw, userRepo)
	listListsUC := usecase.NewListListsUseCase(trelloGw, userRepo)
	selectBoardUC := usecase.NewSelectBoardUseCase(userRepo)
	selectListUC := usecase.NewSelectListUseCase(userRepo)
	registerUserUC := usecase.NewRegisterUserUseCase(userRepo, cfg.TrelloAPIKey)
	connectTrelloUC := usecase.NewConnectTrelloUseCase(userRepo)

	parseIntentUC := usecase.NewParseIntentUseCase(intentChain, userRepo)
	executeActionUC := usecase.NewExecuteActionUseCase(trelloGw, trelloGw, trelloGw, userRepo, taskLogRepo)

	// Adapters
	ctrl := controller.NewTelegramController(
		createTaskUC, parseTaskUC, confirmTaskUC,
		listBoardsUC, listListsUC, selectBoardUC, selectListUC,
		registerUserUC, connectTrelloUC,
		pendingStore,
		parseIntentUC, executeActionUC,
	)
	pres := presenter.NewTelegramPresenter()

	// Delivery
	router := telegram.NewRouter(ctrl, pres, logger)
	bot, err := telegram.NewBot(cfg.TelegramToken, router, logger)
	if err != nil {
		logger.Error("failed to create bot", "error", err)
		os.Exit(1)
	}

	// HTTP Server
	port := cfg.Port
	if port == "" {
		port = "8080"
	}
	mux := http.NewServeMux()
	mux.Handle("/healthz", health.NewHealthHandler(db))

	// Mini App (feature-flagged)
	if cfg.MiniAppEnabled {
		logger.Info("mini app enabled, wiring API routes")

		initDataValidator := miniapp.NewTelegramInitDataValidator(cfg.TelegramToken)
		sessionMgr := miniapp.NewJWTSessionManager(cfg.JWTSecret)

		// Mini App use cases
		authenticateUC := usecase.NewAuthenticateUseCase(initDataValidator, sessionMgr, userRepo)
		getUserSettingsUC := usecase.NewGetUserSettingsUseCase(userRepo)
		listLabelsUC := usecase.NewListLabelsUseCase(trelloGw, userRepo)
		listMembersUC := usecase.NewListMembersUseCase(trelloGw, userRepo)
		listCardsUC := usecase.NewListCardsUseCase(trelloGw, userRepo)
		getCardUC := usecase.NewGetCardUseCase(trelloGw, userRepo)
		createCardFormUC := usecase.NewCreateCardFormUseCase(trelloGw, userRepo, taskLogRepo)
		updateCardUC := usecase.NewUpdateCardUseCase(trelloGw, userRepo)
		deleteCardUC := usecase.NewDeleteCardUseCase(trelloGw, userRepo)
		addCommentUC := usecase.NewAddCommentUseCase(trelloGw, userRepo)

		// Mini App controllers
		miniAppCtrl := controller.NewMiniAppController(
			authenticateUC, getUserSettingsUC,
			listBoardsUC, listListsUC, listLabelsUC,
		)
		cardCtrl := controller.NewMiniAppCardController(
			listMembersUC, listCardsUC, getCardUC,
			createCardFormUC, updateCardUC,
		)
		settingsCtrl := controller.NewMiniAppSettingsController(
			deleteCardUC, addCommentUC,
			selectBoardUC, selectListUC, connectTrelloUC,
		)

		// Embedded frontend
		frontendFS, fsErr := fs.Sub(web.FrontendFS, "dist")
		if fsErr != nil {
			logger.Error("failed to load frontend", "error", fsErr)
			os.Exit(1)
		}

		apiHandler := miniapp.NewAPIRouter(
			miniAppCtrl, cardCtrl, settingsCtrl,
			sessionMgr, logger, frontendFS,
		)
		mux.Handle("/", apiHandler)
	}

	healthServer := &http.Server{Addr: ":" + port, Handler: mux}
	go func() {
		logger.Info("health server started", "port", port)
		if err := healthServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("health server error", "error", err)
		}
	}()

	// Graceful Shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if cfg.TelegramMode == "webhook" && cfg.TelegramWebhookURL != "" {
		go func() {
			if err := bot.StartWebhook(cfg.TelegramWebhookURL, ":8443"); err != nil {
				logger.Error("webhook failed", "error", err)
				os.Exit(1)
			}
		}()
	} else {
		go bot.StartPolling()
	}

	<-ctx.Done()
	logger.Info("shutting down gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := healthServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("health server shutdown error", "error", err)
	}
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
