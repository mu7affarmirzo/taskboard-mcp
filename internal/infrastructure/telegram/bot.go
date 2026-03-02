package telegram

import (
	"log/slog"
	"net/http"

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
	b.logger.Info("bot started in polling mode", "username", b.api.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.api.GetUpdatesChan(u)
	for update := range updates {
		go b.router.Route(b.api, update)
	}
}

// StartWebhook registers a webhook with Telegram and starts an HTTP server
// to receive updates. The webhookURL must be publicly accessible HTTPS.
// The listenAddr is the local address to bind (e.g., ":8443").
func (b *Bot) StartWebhook(webhookURL, listenAddr string) error {
	wh, err := tgbotapi.NewWebhook(webhookURL)
	if err != nil {
		return err
	}
	if _, err := b.api.Request(wh); err != nil {
		return err
	}

	b.logger.Info("bot started in webhook mode",
		"username", b.api.Self.UserName,
		"webhook_url", webhookURL,
		"listen_addr", listenAddr,
	)

	updates := b.api.ListenForWebhook("/webhook")
	go func() {
		if err := http.ListenAndServe(listenAddr, nil); err != nil {
			b.logger.Error("webhook server error", "error", err)
		}
	}()

	for update := range updates {
		go b.router.Route(b.api, update)
	}
	return nil
}
