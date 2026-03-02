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

func BuildConfirmKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Create", "confirm:create"),
			tgbotapi.NewInlineKeyboardButtonData("Edit", "confirm:edit"),
			tgbotapi.NewInlineKeyboardButtonData("Cancel", "confirm:cancel"),
		),
	)
}
