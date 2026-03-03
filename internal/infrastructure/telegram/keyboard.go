package telegram

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"telegram-trello-bot/internal/usecase/port"
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

// webAppButton is a custom InlineKeyboardButton with web_app support.
// The go-telegram-bot-api v5.5.1 SDK does not include the web_app field,
// so we define our own type that serializes the field correctly.
type webAppButton struct {
	Text   string      `json:"text"`
	WebApp *webAppInfo `json:"web_app"`
}

type webAppInfo struct {
	URL string `json:"url"`
}

// webAppKeyboard wraps buttons that include web_app into a format
// compatible with tgbotapi.InlineKeyboardMarkup JSON serialization.
type webAppKeyboard struct {
	InlineKeyboard [][]webAppButton `json:"inline_keyboard"`
}

// BuildWebAppKeyboard creates an inline keyboard with a single WebApp button
// that opens inside Telegram rather than in an external browser.
func BuildWebAppKeyboard(text, url string) webAppKeyboard {
	return webAppKeyboard{
		InlineKeyboard: [][]webAppButton{
			{
				{
					Text:   text,
					WebApp: &webAppInfo{URL: url},
				},
			},
		},
	}
}

// BuildMainMenuKeyboard creates a persistent reply keyboard with the main bot commands.
func BuildMainMenuKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("/boards"),
			tgbotapi.NewKeyboardButton("/app"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("/help"),
			tgbotapi.NewKeyboardButton("/start"),
		),
	)
}

func BuildConfirmKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Create", "confirm:create"),
			tgbotapi.NewInlineKeyboardButtonData("Edit", "confirm:edit"),
			tgbotapi.NewInlineKeyboardButtonData("Cancel", "confirm:cancel"),
		),
	)
}
