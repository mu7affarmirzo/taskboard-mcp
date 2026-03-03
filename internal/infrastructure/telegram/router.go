package telegram

import (
	"context"
	"log/slog"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"telegram-trello-bot/internal/adapter/controller"
	"telegram-trello-bot/internal/adapter/presenter"
	"telegram-trello-bot/internal/usecase/port"
)

type BotSender interface {
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
	Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error)
}

type Router struct {
	ctrl      *controller.TelegramController
	presenter *presenter.TelegramPresenter
	logger    *slog.Logger
}

func NewRouter(ctrl *controller.TelegramController, pres *presenter.TelegramPresenter, logger *slog.Logger) *Router {
	return &Router{ctrl: ctrl, presenter: pres, logger: logger}
}

func (r *Router) Route(api BotSender, update tgbotapi.Update) {
	ctx := context.Background()
	start := time.Now()

	if update.CallbackQuery != nil {
		r.handleCallback(ctx, api, update.CallbackQuery)
		return
	}

	if update.Message == nil {
		return
	}

	chatID := update.Message.Chat.ID
	userID := update.Message.From.ID
	log := r.logger.With("user_id", userID, "chat_id", chatID)

	if update.Message.IsCommand() {
		cmd := update.Message.Command()
		args := update.Message.CommandArguments()
		log.Debug("command received", "command", cmd)
		r.handleCommand(ctx, api, chatID, userID, cmd, args)
		log.Debug("command handled", "command", cmd, "duration_ms", time.Since(start).Milliseconds())
		return
	}

	log.Debug("message received", "text_length", len(update.Message.Text))

	intent, err := r.ctrl.HandleIntent(ctx, userID, update.Message.Text)
	if err != nil {
		log.Error("parse intent failed", "error", err)
		r.sendText(api, chatID, r.presenter.FormatError(err))
		return
	}

	if intent.Action == "create_task" {
		r.ctrl.StorePendingFromIntent(userID, intent)
		kb := BuildConfirmKeyboard()
		r.sendTextWithKeyboard(api, chatID, r.presenter.FormatIntentPreview(intent), kb)
		log.Debug("task preview sent", "title", intent.Title, "duration_ms", time.Since(start).Milliseconds())
		return
	}

	result, err := r.ctrl.HandleExecuteAction(ctx, userID, intent)
	if err != nil {
		log.Error("execute action failed", "action", intent.Action, "error", err)
		r.sendText(api, chatID, r.presenter.FormatError(err))
		return
	}
	r.sendText(api, chatID, r.presenter.FormatActionResult(result))
	log.Debug("action executed", "action", intent.Action, "duration_ms", time.Since(start).Milliseconds())
}

func (r *Router) handleCommand(ctx context.Context, api BotSender, chatID, userID int64, cmd string, args string) {
	switch cmd {
	case "start":
		output, err := r.ctrl.HandleStart(ctx, userID)
		if err != nil {
			r.sendText(api, chatID, r.presenter.FormatError(err))
			return
		}
		r.sendText(api, chatID, r.presenter.FormatWelcome(output))
	case "connect":
		token := strings.TrimSpace(args)
		output, err := r.ctrl.HandleConnectTrello(ctx, userID, token)
		if err != nil {
			r.sendText(api, chatID, r.presenter.FormatError(err))
			return
		}
		if output.Connected {
			r.sendText(api, chatID, r.presenter.FormatTrelloConnected())
		}
	case "help":
		r.sendText(api, chatID, r.presenter.FormatHelp())
	case "boards":
		output, err := r.ctrl.HandleListBoards(ctx, userID)
		if err != nil {
			r.sendText(api, chatID, r.presenter.FormatError(err))
			return
		}
		boardInfos := make([]port.BoardInfo, len(output.Boards))
		for i, b := range output.Boards {
			boardInfos[i] = port.BoardInfo{ID: b.ID, Name: b.Name}
		}
		kb := BuildBoardKeyboard(boardInfos)
		r.sendTextWithKeyboard(api, chatID, r.presenter.FormatBoardList(output), kb)
	default:
		r.sendText(api, chatID, "Unknown command. Try /help")
	}
}

func (r *Router) handleCallback(ctx context.Context, api BotSender, cb *tgbotapi.CallbackQuery) {
	chatID := cb.Message.Chat.ID
	userID := cb.From.ID
	data := cb.Data
	log := r.logger.With("user_id", userID, "chat_id", chatID, "callback", data)
	start := time.Now()

	log.Debug("callback received")

	callback := tgbotapi.NewCallback(cb.ID, "")
	if _, err := api.Request(callback); err != nil {
		log.Error("failed to answer callback", "error", err)
	}

	switch {
	case strings.HasPrefix(data, "board:"):
		boardID := strings.TrimPrefix(data, "board:")
		if err := r.ctrl.HandleSelectBoard(ctx, userID, boardID); err != nil {
			r.sendText(api, chatID, r.presenter.FormatError(err))
			return
		}
		listOutput, err := r.ctrl.HandleListLists(ctx, userID, boardID)
		if err != nil {
			r.sendText(api, chatID, r.presenter.FormatError(err))
			return
		}
		listInfos := make([]port.ListInfo, len(listOutput.Lists))
		for i, l := range listOutput.Lists {
			listInfos[i] = port.ListInfo{ID: l.ID, Name: l.Name}
		}
		kb := BuildListKeyboard(listInfos)
		r.sendTextWithKeyboard(api, chatID, r.presenter.FormatBoardSelected(boardID), kb)

	case strings.HasPrefix(data, "list:"):
		listID := strings.TrimPrefix(data, "list:")
		if err := r.ctrl.HandleSelectList(ctx, userID, listID); err != nil {
			r.sendText(api, chatID, r.presenter.FormatError(err))
			return
		}
		r.sendText(api, chatID, r.presenter.FormatListSelected(listID))

	case data == "confirm:create":
		output, err := r.ctrl.HandleConfirmTask(ctx, userID)
		if err != nil {
			log.Error("confirm task failed", "error", err)
			r.sendText(api, chatID, r.presenter.FormatError(err))
			return
		}
		r.sendText(api, chatID, r.presenter.FormatTaskCreated(output))
		log.Info("task created", "title", output.TaskTitle, "duration_ms", time.Since(start).Milliseconds())

	case data == "confirm:edit":
		r.ctrl.HandleCancelTask(userID)
		r.sendText(api, chatID, "Send your edited message:")

	case data == "confirm:cancel":
		r.ctrl.HandleCancelTask(userID)
		r.sendText(api, chatID, "Task cancelled.")

	default:
		log.Warn("unknown callback data")
	}

	log.Debug("callback handled", "duration_ms", time.Since(start).Milliseconds())
}

func (r *Router) sendText(api BotSender, chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	if _, err := api.Send(msg); err != nil {
		r.logger.Error("failed to send message", "error", err, "chat_id", chatID)
	}
}

func (r *Router) sendTextWithKeyboard(api BotSender, chatID int64, text string, keyboard tgbotapi.InlineKeyboardMarkup) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard
	if _, err := api.Send(msg); err != nil {
		r.logger.Error("failed to send message with keyboard", "error", err, "chat_id", chatID)
	}
}
